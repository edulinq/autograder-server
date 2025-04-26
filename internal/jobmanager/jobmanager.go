package jobmanager

import (
	"context"
	"errors"
	"fmt"

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
// It supports the following optional functionality:
//   - synchronization
//   - context cancellation
//   - record retrieval
//   - record storage
//   - record removal from storage
//   - asynchronous output processing
//
// Given job options, input items, and a work function,
// Job will process a list of input items through the work func in a parallel pool to produce a JobOutput.
// If a lock key is provided, Job will block on that key before running preventing overuse of resources or conflicts between the same Job.
// Provide an input key generation function to synchronize at the input item level.
type Job[InputType any, OutputType any] struct {
	// The user level options for a Job.
	*JobOptions

	// Return a copy of the incomplete result.
	// The final output will not be accessible.
	// If WaitForCompletion is true, this option is ignored.
	ReturnIncompleteResults bool

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
	RemoveFunc func([]InputType) error

	// A function to transform work items into results.
	// Returns a result, the time of computation, and an error.
	WorkFunc func(InputType) (OutputType, error)

	// An optional function to get the locking key for an individual work item.
	WorkItemKeyFunc func(InputType) string

	// An optional function to process final JobOutput.
	OnComplete func(JobOutput[InputType, OutputType])
}

// JobOutput is the result of a call to Job.Run().
// It contains a system error, a map of work errors, the result items,
// the remaining items, the total run time, and a channel that signals execution is complete.
// If the context is cacelled while running a job,
// the returned JobOutput will be empty.
type JobOutput[InputType any, OutputType any] struct {
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

	// The total computation time spent in WorkFunc().
	// This time does not include time spent waiting for locks or retrieving stored items.
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
// When not waiting for completion, the JobOutput may still be modified until the Done channel is closed.
// Returning a copy of results gives immediate access to stable results but the final results will not be accessible later.
func (this *Job[InputType, OutputType]) Run() *JobOutput[InputType, OutputType] {
	done := make(chan any)

	output := JobOutput[InputType, OutputType]{
		Done:           done,
		ResultItems:    make([]OutputType, 0, len(this.WorkItems)),
		RemainingItems: this.WorkItems,
		RunTime:        0,
		WorkErrors:     make(map[int]error, 0),
	}

	err := this.Validate()
	if err != nil {
		output.Error = fmt.Errorf("Failed to validate job: '%w'.", err)

		close(done)
		return &output
	}

	// If we are overwriting records, remove all the old records.
	if this.OverwriteRecords && !this.DryRun && this.RemoveFunc != nil {
		err = this.RemoveFunc(this.WorkItems)
		if err != nil {
			output.Error = fmt.Errorf("Failed to remove items: '%w'.", err)

			close(done)
			return &output
		}
	}

	// Query for stored records.
	if !this.OverwriteRecords && this.RetrieveFunc != nil {
		output.ResultItems, output.RemainingItems, err = this.RetrieveFunc(this.WorkItems)
		if err != nil {
			output.Error = fmt.Errorf("Failed to retrieve items: '%w'.", err)

			close(done)
			return &output
		}
	}

	if this.WaitForCompletion {
		this.run(&output, true)

		close(done)
	} else if this.ReturnIncompleteResults {
		go func() {
			backgroundDone := make(chan any)
			defer close(backgroundDone)

			backgroundOutput := &JobOutput[InputType, OutputType]{
				Done:           backgroundDone,
				ResultItems:    make([]OutputType, 0),
				RemainingItems: output.RemainingItems,
				RunTime:        0,
				WorkErrors:     make(map[int]error, 0),
			}

			this.run(backgroundOutput, false)
			if backgroundOutput.Error != nil {
				log.Error("Failure while running asynchronous job.", backgroundOutput.Error)
			}
		}()

		close(done)
	} else {
		go func() {
			this.run(&output, true)
			if output.Error != nil {
				log.Error("Failure while running asynchronous job.", output.Error)
			}

			close(done)
		}()
	}

	// If the context was canceled during execution, return immediately.
	if this.Context.Err() != nil {
		return nil
	}

	return &output
}

// Job.run() processes the remaining items and updates the partial job output.
// When update output is false, Job.run() will not update the results to reduce memory usage.
// However, the run time and error will be updated for stats and logging purposes.
func (this *Job[InputType, OutputType]) run(output *JobOutput[InputType, OutputType], updateOutput bool) {
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
	if !noLockWait && this.RetrieveFunc != nil {
		partialResults := []OutputType{}
		remainingItems := []InputType{}
		partialResults, remainingItems, err = this.RetrieveFunc(output.RemainingItems)
		if err != nil {
			output.Error = fmt.Errorf("Failed to re-check record storage before run: '%w'.", err)
			return
		}

		output.RemainingItems = remainingItems

		if updateOutput {
			// Collect the partial records from storage.
			output.ResultItems = append(output.ResultItems, partialResults...)
		}
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

		// Check storage for a record.
		if this.RetrieveFunc != nil {
			results, _, err := this.RetrieveFunc([]InputType{workItem})
			if err != nil {
				return PoolResult{
					Input: workItem,
					Error: fmt.Errorf("Failed to retrieve item '%v': '%w'.", workItem, err),
				}, nil
			}

			if len(results) == 1 {
				return PoolResult{workItem, results[0], 0, nil}, nil
			}
		}

		// Nothing cached, compute the job.
		startTime := timestamp.Now()

		result, err := this.WorkFunc(workItem)
		if err != nil {
			return PoolResult{
				Input: workItem,
				Error: fmt.Errorf("Failed to perform individual work on item '%v': '%w'.", workItem, err),
			}, nil
		}

		runTime := (timestamp.Now() - startTime).ToMSecs()

		// Store the result.
		if !this.DryRun && this.Context.Err() == nil && this.StoreFunc != nil {
			err = this.StoreFunc([]OutputType{result})
			if err != nil {
				return PoolResult{
					Input: workItem,
					Error: fmt.Errorf("Failed to store result for item '%v': '%w'.", workItem, err),
				}, nil
			}
		}

		return PoolResult{workItem, result, runTime, nil}, nil
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
			// Always update the error for error logging purposes.
			output.Error = errors.Join(output.Error, poolResult.Error)

			if updateOutput {
				output.WorkErrors[len(output.RemainingItems)] = poolResult.Error
				output.RemainingItems = append(output.RemainingItems, poolResult.Input)
			}
		} else {
			// Always update the run time for stats purposes.
			output.RunTime += poolResult.RunTime

			if updateOutput {
				output.ResultItems = append(output.ResultItems, poolResult.Result)
			}
		}
	}

	if this.OnComplete != nil {
		this.OnComplete(*output)
	}
}
