package jobmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// JobOptions contains user level options for a Job.
// These options allow for optional context cancellation of the Job.
type JobOptions struct {
	// Don't save anything.
	DryRun bool `json:"dry-run"`

	// Remove any existing records before running the job.
	OverwriteRecords bool `json:"overwrite-records"`

	// Wait for the entire job to complete and return all results.
	WaitForCompletion bool `json:"wait-for-completion"`

	// A context that can be used to cancel the job.
	Context context.Context `json:"-"`
}

// Job provides system level customization of the job's execution.
// It supports optional record retrieval, record storage, record removal from storage, synchronization, and context cancellation.
// Given job options, input items, an optional record lookup/removal function, and a work function,
// Job will process a list of input items through the work func in a parallel pool to produce a JobOutput.
// If a lock key is provided, Job will block on that key before running preventing overuse of resources or conflicts between the same Job.
// Provide an input key generation function to synchronize at the input item level.
type Job[InputType any, OutputType any] struct {
	// The user level options for a Job.
	*JobOptions

	// The number of workers in the parallel pool.
	PoolSize int

	// An optional key to lock on.
	LockKey string

	// A sorted list of items that need work.
	WorkItems []InputType

	// An optional function to retrieve existing records that should not be processed.
	// Returns a list of processed records, remaining items, and an error.
	RetrieveFunc func([]InputType) ([]OutputType, []InputType, error)

	// An optional function to store the result.
	StoreFunc func([]OutputType) error

	// An optional function to remove existing records from storage.
	// TODO: Find a way to not remove when on DryRun (may need to modify retrieve func).
	RemoveStorageFunc func([]InputType) error

	// A function to transform work items into results.
	// Returns a result, the time of computation, and an error.
	WorkFunc func(InputType) (OutputType, error)

	// An optional function to get the locking key for an individual work item.
	WorkItemKeyFunc func(InputType) string
}

// JobOutput is the result of a call to Job.Run().
// It contains a system error, a map of work errors, the result items,
// the remaining items, the total run time, and a channel that signals execution is complete.
// If the context is cacelled while running a job,
// the returned JobOutput will be empty.
type JobOutput[InputType any, OutputType any] struct {
	// TODO: Can lock shared JobOutput and provide interface methods to get data (that handles the locking).
	// An error for Job.Run().
	Error error

	// A map of errors encountered while running the job.
	// The key of the error will be the index of the item
	// that failed in RemainingItems.
	WorkErrors map[int]error

	// The list of results.
	ResultItems []OutputType

	// The list of tems that still need work.
	RemainingItems []InputType

	// The total computation time.
	RunTime int64

	// Signals the job is complete.
	Done <-chan any
}

func (this *Job[InputType, OutputType]) Validate() error {
	if this == nil {
		return fmt.Errorf("Job is nil.")
	}

	if this.WorkFunc == nil {
		return fmt.Errorf("Job cannot have a nil work function.")
	}

	if this.PoolSize <= 0 {
		return fmt.Errorf("Pool size must be positive, got %d.", this.PoolSize)
	}

	return this.JobOptions.Validate()
}

func (this *JobOptions) Validate() error {
	if this == nil {
		return fmt.Errorf("Job options are nil.")
	}

	if this.Context == nil {
		this.Context = context.Background()
	}

	return nil
}

// Given a customized Job, Job.Run() processes input items in a parallel pool of workers.
// Returns the collected results in a JobOutput.
// If the context is canceled during execution, returns nil.
// When not waiting for completion, Job.JobOutput will be populated with the results when the JobOutput.Done channel is closed.
func (this *Job[InputType, OutputType]) Run() *JobOutput[InputType, OutputType] {
	err := this.Validate()
	if err != nil {
		return &JobOutput[InputType, OutputType]{
			Error: fmt.Errorf("Failed to validate job: '%v'.", err),
		}
	}

	done := make(chan any)

	output := JobOutput[InputType, OutputType]{
		Done:           done,
		ResultItems:    make([]OutputType, 0, len(this.WorkItems)),
		RemainingItems: this.WorkItems,
		RunTime:        0,
		WorkErrors:     make(map[int]error, 0),
	}

	// If we are overwriting records, remove all the old records.
	if this.OverwriteRecords && !this.DryRun && this.RemoveStorageFunc != nil {
		output.Error = this.RemoveStorageFunc(this.WorkItems)
		if output.Error != nil {
			return &output
		}
	}

	// Query for stored records.
	if !this.OverwriteRecords && this.RetrieveFunc != nil {
		output.ResultItems, output.RemainingItems, output.Error = this.RetrieveFunc(this.WorkItems)
		if output.Error != nil {
			return &output
		}
	}

	if this.WaitForCompletion {
		this.run(&output)

		close(done)

		// If the context was canceled during execution, return immediately.
		if this.Context.Err() != nil {
			return nil
		}
	} else {
		// TODO: this.(non embedded)JobOutput.Copy()
		// return copy
		go func() {
			// TEST
			// Make a new job output here.
			// Only takes remaining items.
			this.run(&output)
			if output.Error != nil {
				log.Error("Failure while running asynchronous job: '%v'.", output.Error)
			}

			close(done)
		}()
	}

	return &output
}

// TODO: Can pass in a JobOutput pointer to modify.
// If it's nil, don't save to results. (only errors)
func (this *Job[InputType, OutputType]) run(output *JobOutput[InputType, OutputType]) {
	if len(output.RemainingItems) == 0 {
		return
	}

	noLockWait := true
	if this.LockKey != "" {
		noLockWait = lockmanager.Lock(this.LockKey)
		defer lockmanager.Unlock(this.LockKey)
	}

	// The context has been canceled while waiting for a lock, abandon this job.
	if this.Context.Err() != nil {
		return
	}

	var err error = nil

	// If we had to wait for the lock, then check again for stored records.
	if !noLockWait && !this.OverwriteRecords && this.RetrieveFunc != nil {
		partialResults := []OutputType{}
		partialResults, output.RemainingItems, err = this.RetrieveFunc(output.RemainingItems)
		if err != nil {
			output.Error = fmt.Errorf("Failed to re-check record storage before run: '%w'.", err)
			return
		}

		// Collect the partial records from storage.
		output.ResultItems = append(output.ResultItems, partialResults...)
	}

	if len(output.RemainingItems) == 0 {
		return
	}

	type PoolResult struct {
		Input   InputType
		Result  OutputType
		RunTime int64
		Error   error
	}

	poolResults, _, err := util.RunParallelPoolMap(this.PoolSize, output.RemainingItems, this.Context, func(workItem InputType) (PoolResult, error) {
		workItemKey := ""
		if this.WorkItemKeyFunc != nil {
			workItemKey = this.WorkItemKeyFunc(workItem)
		}

		// Optionally lock this id so we don't work on an item multiple times.
		if workItemKey != "" {
			lockmanager.Lock(workItemKey)
			defer lockmanager.Unlock(workItemKey)
		}

		if this.Context.Err() != nil {
			return PoolResult{}, nil
		}

		startTime := timestamp.Now()

		var errs error = err

		result, err := this.WorkFunc(workItem)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to perform individual work on item %v: '%w'.", workItem, err))
		}

		runTime := (timestamp.Now() - startTime).ToMSecs()
		// Standardize the run time so tests can check for exact matches.
		if config.UNIT_TESTING_MODE.Get() {
			runTime = 1
		}

		// We can separate the semantic options from the save func and retrieve func.
		// Store the result.
		if !this.DryRun && this.Context.Err() == nil && this.StoreFunc != nil {
			err = this.StoreFunc([]OutputType{result})
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("Failed to store result for item '%v': '%w'.", workItem, err))
			}
		}

		return PoolResult{workItem, result, runTime, errs}, nil
	})

	if err != nil {
		output.Error = fmt.Errorf("Failed to run job in a parallel pool: '%w'.", err)
		return
	}

	// If the job was canceled, exit right away.
	if this.Context.Err() != nil {
		return
	}

	output.RemainingItems = []InputType{}

	for _, poolResult := range poolResults {
		if poolResult.Error != nil {
			output.WorkErrors[len(output.RemainingItems)] = poolResult.Error
			output.RemainingItems = append(output.RemainingItems, poolResult.Input)
			output.Error = errors.Join(output.Error, poolResult.Error)
		} else {
			output.ResultItems = append(output.ResultItems, poolResult.Result)
			output.RunTime += poolResult.RunTime
		}
	}
}
