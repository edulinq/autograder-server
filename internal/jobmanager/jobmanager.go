package jobmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

type Job[InputType comparable, OutputType any] struct {
	JobOptions

	JobOutput[InputType, OutputType]

	// A sorted list of items that need work.
	WorkItems []InputType

	// A function to retrieve existing records.
	// Returns a list of processed records, items that need work, and an error.
	RetrieveFunc func([]InputType) ([]OutputType, []InputType, error) `json:"-"`

	// A function to remove existing records from storage.
	RemoveStorageFunc func([]InputType) error `json:"-"`

	// A function to transform work items into results.
	// Returns a result, the time of computation, and an error.
	WorkFunc func(InputType) (OutputType, int64, error) `json:"-"`

	// An internal channel to signal the job is complete.
	done chan struct{} `json:"-"`
}

type JobOutput[InputType comparable, OutputType any] struct {
	// A list of errors encountered while running the job.
	// The index of the error will match the index of the item
	// that failed in RemainingItems.
	Errors []error

	// The list of results.
	ResultItems []OutputType

	// The list of tems that still need work.
	RemainingItems []InputType

	// The total computation time.
	RunTime int64

	// Signals the job is complete.
	Done <-chan struct{} `json:"-"`
}

type JobOptions struct {
	// Replace any records currently in storage,
	// and do not retrieve any records (when not waiting for completion).
	OverwriteRecords bool `json:"overwrite-records"`

	// Wait for the entire job to complete and return all results.
	WaitForCompletion bool `json:"wait-for-completion"`

	// A context that can be used to cancel the job.
	Context context.Context `json:"-"`

	// If true, do not swap the context to the background context when running.
	// By default (when this is false), the context will be swapped to the background context when !WaitForCompletion.
	// The swap is so that jobs do not get canceled when an HTTP request is complete.
	// Setting this true is useful for testing (as one round of job tests can be wrapped up).
	RetainOriginalContext bool `json:"-"`

	// The number of workers in the parallel pool.
	PoolSize int `json:"-"`

	// An optional key to lock on.
	LockKey string `json:"-"`
}

func (this *Job[InputType, OutputType]) Validate() error {
	if this == nil {
		return fmt.Errorf("Job is nil.")
	}

	if this.WorkFunc == nil {
		return fmt.Errorf("Job cannot have a nil work function.")
	}

	return this.JobOptions.Validate()
}

func (this *JobOptions) Validate() error {
	if this == nil {
		return fmt.Errorf("Job options are nil.")
	}

	if this.PoolSize <= 0 {
		return fmt.Errorf("Pool size must be positive, got %d.", this.PoolSize)
	}

	if this.Context == nil {
		this.Context = context.Background()
	}

	if !this.WaitForCompletion && !this.RetainOriginalContext {
		this.Context = context.Background()
	}

	return nil
}

// Job.Run() executes a potentially long-running job over a list of input items.
// It supports optional record retrieval, record removal from storage, synchronization, and context cancellation.
// Given job options, input items, an optional record lookup/removal function, and a work function,
// Job.Run() processes each item in a parallel pool.
// If a lock key is provided, Job.Run() will block on that key for the duration of the method.
// This prevents multiple Job.Run() invocations of the same type from overusing resources or conflicting with each other.
// Returns the result list, number of remaining items, total run time, and an error.
// If the context is canceled, returns nil, 0, 0, nil.
func (this *Job[InputType, OutputType]) Run() (JobOutput[InputType, OutputType], error) {
	err := this.Validate()
	if err != nil {
		return JobOutput[InputType, OutputType]{}, fmt.Errorf("Failed to validate job: '%v'.", err)
	}

	this.ResultItems = make([]OutputType, 0, len(this.WorkItems))
	this.RemainingItems = this.WorkItems
	this.RunTime = 0

	// If we are not overwriting records, query for stored records.
	if !this.OverwriteRecords && this.RetrieveFunc != nil {
		this.ResultItems, this.RemainingItems, err = this.RetrieveFunc(this.WorkItems)
		if err != nil {
			return JobOutput[InputType, OutputType]{}, err
		}
	}

	if this.done == nil {
		this.done = make(chan struct{})
		this.Done = this.done
	}

	if this.WaitForCompletion {
		err = this.run()
		if err != nil {
			return JobOutput[InputType, OutputType]{}, err
		}

		// If the context was canceled during execution, return immediately.
		if this.Context.Err() != nil {
			return JobOutput[InputType, OutputType]{}, nil
		}

		select {
		case <-this.done:
		default:
			close(this.done)
		}
	} else {
		go func() {
			err = this.run()
			if err != nil {
				log.Error("Failure while running asynchronous job.")
			}

			select {
			case <-this.done:
			default:
				close(this.done)
			}
		}()
	}

	return this.JobOutput, nil
}

func (this *Job[InputType, OutputType]) run() error {
	err := this.Validate()
	if err != nil {
		return fmt.Errorf("Failed to validate job: '%v'.", err)
	}

	if len(this.RemainingItems) == 0 {
		return nil
	}

	noLockWait := true
	if this.LockKey != "" {
		noLockWait = lockmanager.Lock(this.LockKey)
		defer lockmanager.Unlock(this.LockKey)
	}

	// The context has been canceled while waiting for a lock, abandon this job.
	if this.Context.Err() != nil {
		return nil
	}

	// If we had to wait for the lock, then check again for stored records.
	if !noLockWait && this.RetrieveFunc != nil {
		var partialResults []OutputType = nil
		partialResults, this.RemainingItems, err = this.RetrieveFunc(this.RemainingItems)
		if err != nil {
			return fmt.Errorf("Failed to re-check record storage before run: '%w'.", err)
		}

		// Collect the partial records from storage.
		this.ResultItems = append(this.ResultItems, partialResults...)
	}

	if len(this.RemainingItems) == 0 {
		return nil
	}

	// If we are overwriting records, then remove all the old records.
	if this.OverwriteRecords && this.RemoveStorageFunc != nil {
		processedItems := getProcessedItems(this.WorkItems, this.RemainingItems)
		err = this.RemoveStorageFunc(processedItems)
		if err != nil {
			return fmt.Errorf("Failed to remove old job cache entries: '%w'.", err)
		}
	}

	type PoolResult struct {
		Input   InputType
		Result  OutputType
		RunTime int64
		Error   error
	}

	poolResults, _, err := util.RunParallelPoolMap(this.PoolSize, this.RemainingItems, this.Context, func(workItem InputType) (PoolResult, error) {
		result, runTime, err := this.WorkFunc(workItem)
		if err != nil {
			err = fmt.Errorf("Failed to perform individual work on item %v: '%w'.", workItem, err)
		}

		return PoolResult{workItem, result, runTime, err}, nil
	})

	if err != nil {
		return fmt.Errorf("Failed to run job in a parallel pool: '%w'.", err)
	}

	// If the job was canceled, exit right away.
	if this.Context.Err() != nil {
		return nil
	}

	this.RunTime = 0
	this.Errors = nil
	this.RemainingItems = []InputType{}

	for _, poolResult := range poolResults {
		if poolResult.Error != nil {
			this.Errors = append(this.Errors, poolResult.Error)
			this.RemainingItems = append(this.RemainingItems, poolResult.Input)
		} else {
			this.ResultItems = append(this.ResultItems, poolResult.Result)
			this.RunTime += poolResult.RunTime
		}
	}

	return errors.Join(this.Errors...)
}

func getProcessedItems[InputType comparable](allItems []InputType, remainingItems []InputType) []InputType {
	seenItems := make(map[InputType]bool, len(remainingItems))
	for _, item := range remainingItems {
		seenItems[item] = true
	}

	results := make([]InputType, 0, len(allItems)-len(remainingItems))

	for _, item := range allItems {
		_, ok := seenItems[item]
		if !ok {
			results = append(results, item)
		}
	}

	return results
}
