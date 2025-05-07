package jobmanager

import (
	"context"
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
// Job will invoke the work func on each input item in a parallel pool to produce a JobOutput.
// If a lock key is provided, Job will block on that key before running.
// The locking behavior prevents overuse of resources and conflicts between the same Job.
// Provide an input key generation function to synchronize at the input item level in addition to the job level.
type Job[InputType comparable, OutputType any] struct {
	// The user level options for a Job.
	*JobOptions

	// Instead of adding to the results as they are processed,
	// return a copy of the potentially incomplete results fetched before the job starts.
	// If WaitForCompletion is true, this option is ignored.
	ReturnIncompleteResults bool

	// The number of workers in the parallel pool.
	PoolSize int

	// An optional key to lock the job on.
	// The job will block until it acquires the lock.
	// Provide a WorkItemKeyFunc to synchronize individual work items.
	LockKey string

	// A list of items that need work.
	WorkItems []InputType

	// An optional function to retrieve existing records that should not be processed.
	// Returns a list of processed records, remaining items, and an error.
	// E.g. This function could be used to retrieve completed items from a cache.
	RetrieveFunc func([]InputType) (map[InputType]OutputType, error)

	// An optional function to store the result.
	// Use this function to cache the result of the work function to reduce future computation.
	StoreFunc func([]OutputType) error

	// An optional function to remove existing records from storage.
	RemoveFunc func([]InputType) error

	// A function to transform work items into results.
	// Returns a result and an error.
	WorkFunc func(InputType) (OutputType, error)

	// An optional function to get the locking key for an individual work item.
	WorkItemKeyFunc func(InputType) string

	// An optional function to process the final JobOutput upon successful completion.
	OnSuccess func(JobOutput[InputType, OutputType])
}

// JobOutput is the result of a call to Job.Run().
// It contains a system error, a map of work errors, the result items,
// the remaining items, the total run time, and a channel that signals execution is complete.
// System errors are held in Error, whereas errors during work are included in WorkErrors.
// If there are work errors there will always be a system error, but the converse is not always true.
// Always wait for the Done channel to be closed before handling JobOutput.
type JobOutput[InputType comparable, OutputType any] struct {
	// Signals the job was canceled during execution.
	Canceled bool

	// An error for Job.Run().
	Error error

	// A map of errors encountered while running the job.
	// The key of the error will be the index of the item
	// that failed in RemainingItems.
	WorkErrors map[InputType]error

	// The map of results.
	ResultItems map[InputType]OutputType

	// The list of items that still need work.
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
// When not waiting for completion, the JobOutput may still be modified until the Done channel is closed.
// Returning a copy of results gives immediate access to stable results but the final results will not be accessible later.
// Returns nil if the context is canceled during execution.
func (this *Job[InputType, OutputType]) Run() *JobOutput[InputType, OutputType] {
	// Run closes the channel to signal the returned JobOutput is safe to handle.
	// The channel is closed by this thread except when !WaitForCompletion and !ReturnIncompleteResults.
	done := make(chan any)

	output := JobOutput[InputType, OutputType]{
		Done:           done,
		ResultItems:    make(map[InputType]OutputType, len(this.WorkItems)),
		RemainingItems: this.WorkItems,
		RunTime:        0,
		WorkErrors:     make(map[InputType]error, 0),
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
		output.ResultItems, err = this.RetrieveFunc(this.WorkItems)
		if err != nil {
			output.Error = fmt.Errorf("Failed to retrieve items: '%w'.", err)

			close(done)
			return &output
		}

		output.RemainingItems = getRemainingItems(this.WorkItems, output.ResultItems)
	}

	if this.WaitForCompletion {
		this.run(&output, true)

		close(done)
	} else if this.ReturnIncompleteResults {
		backgroundDone := make(chan any)

		backgroundOutput := &JobOutput[InputType, OutputType]{
			Done:           backgroundDone,
			ResultItems:    make(map[InputType]OutputType, len(output.RemainingItems)),
			RemainingItems: output.RemainingItems,
			RunTime:        0,
			WorkErrors:     make(map[InputType]error, 0),
		}

		go func() {
			defer close(backgroundDone)

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
		return &output
	}

	return &output
}

// Job.run() processes the remaining items and updates the partial job output.
// If Job.Run() returned incomplete results, the JobOutput will not be accessible.
// In this case, Job.run() will not not update the results to reduce memory usage.
// However, the run time and error will be updated for stats and logging purposes.
// The parameter updateOutput signals whether Job.run() will add results to JobOutput.
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
		output.Error = fmt.Errorf("Job was canceled: '%v'.", this.Context.Err())
		output.Canceled = true
		return
	}

	// If we had to wait for the lock, then check again for stored records.
	if !noLockWait && this.RetrieveFunc != nil {
		partialResults, err := this.RetrieveFunc(output.RemainingItems)
		if err != nil {
			output.Error = fmt.Errorf("Failed to re-check record storage before run: '%w'.", err)
			return
		}

		output.RemainingItems = getRemainingItems(output.RemainingItems, partialResults)

		if updateOutput {
			// Collect the partial records from storage.
			for input, result := range partialResults {
				output.ResultItems[input] = result
			}
		}
	}

	if len(output.RemainingItems) == 0 {
		return
	}

	err := this.runParallelPoolMap(output, updateOutput)
	if err != nil {
		output.Error = fmt.Errorf("Failed to run job in a parallel pool: '%w'.", err)
		return
	}

	// If the job was canceled, return after collecting partial results.
	if this.Context.Err() != nil {
		output.Error = fmt.Errorf("Job was canceled: '%v'.", this.Context.Err())
		output.Canceled = true
		return
	}

	if len(output.WorkErrors) > 0 {
		output.Error = fmt.Errorf("Failed to complete work for '%d' items.", len(output.WorkErrors))
		return
	}

	if this.OnSuccess != nil {
		this.OnSuccess(*output)
	}
}

func (this *Job[InputType, OutputType]) runParallelPoolMap(output *JobOutput[InputType, OutputType], updateOutput bool) error {
	type resultItem struct {
		Output  OutputType
		RunTime int64
	}

	results, err := util.RunParallelPoolMap(this.PoolSize, output.RemainingItems, this.Context, func(workItem InputType) (resultItem, error) {
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
			return resultItem{}, nil
		}

		// Check storage for a record.
		if this.RetrieveFunc != nil {
			results, err := this.RetrieveFunc([]InputType{workItem})
			if err != nil {
				return resultItem{}, fmt.Errorf("Failed to retrieve item '%v': '%w'.", workItem, err)
			}

			result, ok := results[workItem]
			if ok {
				return resultItem{result, 0}, nil
			}
		}

		// Nothing cached, compute the job.
		startTime := timestamp.Now()

		result, err := this.WorkFunc(workItem)
		if err != nil {
			return resultItem{}, fmt.Errorf("Failed to perform individual work on item '%v': '%w'.", workItem, err)
		}

		if this.Context.Err() != nil {
			return resultItem{}, nil
		}

		runTime := (timestamp.Now() - startTime).ToMSecs()

		// Store the result.
		if !this.DryRun && this.StoreFunc != nil {
			err = this.StoreFunc([]OutputType{result})
			if err != nil {
				return resultItem{}, fmt.Errorf("Failed to store result for item '%v': '%w'.", workItem, err)
			}
		}

		return resultItem{result, runTime}, nil
	})

	output.RemainingItems = []InputType{}

	// The job was canceled so return without collecting results.
	if results.Canceled {
		output.Canceled = true
		return nil
	}

	results.IsDone()

	output.WorkErrors = results.WorkErrors

	for input, result := range results.Results {
		_, ok := results.WorkErrors[input]
		if ok {
			continue
		}

		// Always update the run time for stats purposes.
		output.RunTime += result.RunTime

		if updateOutput {
			output.ResultItems[input] = result.Output
		}
	}

	return err
}

func getRemainingItems[InputType comparable, OutputType any](workItems []InputType, results map[InputType]OutputType) []InputType {
	remainingItems := []InputType{}

	for _, workItem := range workItems {
		_, ok := results[workItem]
		if !ok {
			remainingItems = append(remainingItems, workItem)
		}
	}

	return remainingItems
}
