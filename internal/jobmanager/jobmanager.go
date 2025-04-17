package jobmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

type JobOptions struct {
	// Don't save anything.
	DryRun bool `json:"dry-run"`

	// Replace any entries currently in the cache,
	// and do not return any cached entries (when not waiting for completion).
	OverwriteCache bool `json:"overwrite-cache"`

	// Wait for the entire job to complete and return all results.
	WaitForCompletion bool `json:"wait-for-completion"`

	// A context that can be used to cancel the job.
	Context context.Context `json:"-"`

	// If true, do not swap the context to the background context when running.
	// By default (when this is false), the context will be swapped to the background context when !WaitForCompletion.
	// The swap is so that jobs do not get canceled when an HTTP request is complete.
	// Setting this true is useful for testing (as one round of job tests can be wrapped up).
	RetainOriginalContext bool `json:"-"`

	PoolSize int `json:"-"`

	LockKey string `json:"-"`
}

func (this *JobOptions) Validate() error {
	if this == nil {
		return fmt.Errorf("Job options are nil.")
	}

	if this.LockKey == "" {
		return fmt.Errorf("Cannot have an empty lock key.")
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

// TODO: expand usage
// Returns runtime, output, numRemaining, err
func RunJob[InputType any, OutputType any](options *JobOptions, workItems []InputType, cacheFunc func([]InputType) ([]OutputType, []InputType, error), removeCacheFunc func([]InputType) error, workFunc func(InputType) (OutputType, int64, error)) ([]OutputType, int, int64, error) {
	err := options.Validate()
	if err != nil {
		return []OutputType{}, len(workItems), 0, fmt.Errorf("Failed to validate job options: '%v'.", err)
	}

	completeItems := make([]OutputType, 0)
	remainingItems := workItems

	// If we are not overwriting the cache, query for cached results.
	if !options.OverwriteCache && cacheFunc != nil {
		completeItems, remainingItems, err = cacheFunc(workItems)
		if err != nil {
			return []OutputType{}, len(workItems), 0, err
		}
	}

	runTime := int64(0)
	results := make([]OutputType, 0)

	if options.WaitForCompletion {
		results, runTime, err = runJob(options, remainingItems, cacheFunc, removeCacheFunc, workFunc)
		if err != nil {
			return []OutputType{}, len(workItems), 0, err
		}

		completeItems = append(completeItems, results...)
		remainingItems = nil
	} else {
		go func() {
			_, _, err := runJob(options, remainingItems, cacheFunc, removeCacheFunc, workFunc)
			if err != nil {
				log.Error("Failure while running asynchronous job.")
			}
		}()
	}

	return completeItems, len(remainingItems), runTime, nil
}

func runJob[InputType any, OutputType any](options *JobOptions, workItems []InputType, cacheFunc func([]InputType) ([]OutputType, []InputType, error), removeCacheFunc func([]InputType) error, workFunc func(InputType) (OutputType, int64, error)) ([]OutputType, int64, error) {
	if len(workItems) == 0 {
		return nil, 0, nil
	}

	err := options.Validate()
	if err != nil {
		return []OutputType{}, 0, fmt.Errorf("Failed to validate job options: '%v'.", err)
	}

	noLockWait := lockmanager.Lock(options.LockKey)
	defer lockmanager.Unlock(options.LockKey)

	// The context has been canceled while waiting for a lock, abandon this job.
	if options.Context.Err() != nil {
		return nil, 0, nil
	}

	results := make([]OutputType, 0, len(workItems))

	// If we had to wait for the lock, then check again for cached results.
	// If there are multiple requests queued up,
	// it will be faster to do a bulk check for cached results instead of checking each one individually.
	if !noLockWait && cacheFunc != nil {
		var partialResults []OutputType = nil
		partialResults, workItems, err = cacheFunc(workItems)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to re-check result cache before run: '%w'.", err)
		}

		// Collect the partial results from the cache.
		results = append(results, partialResults...)
	}

	if len(workItems) == 0 {
		return results, 0, nil
	}

	// If we are overwriting the cache, then remove all the old entries.
	if options.OverwriteCache && removeCacheFunc != nil && !options.DryRun {
		err = removeCacheFunc(workItems)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to remove old job cache entries: '%w'.", err)
		}
	}

	type PoolResult struct {
		Result  OutputType
		RunTime int64
		Error   error
	}

	poolResults, _, err := util.RunParallelPoolMap(options.PoolSize, workItems, options.Context, func(workItem InputType) (PoolResult, error) {
		result, runTime, err := workFunc(workItem)
		if err != nil {
			err = fmt.Errorf("Failed to perform individual work on item %v: '%w'.", workItem, err)
		}

		return PoolResult{result, runTime, err}, nil
	})

	if err != nil {
		return nil, 0, fmt.Errorf("Failed to run job in a parallel pool: '%w'.", err)
	}

	// If the job was canceled, exit right away.
	if options.Context.Err() != nil {
		return nil, 0, nil
	}

	var errs error = nil
	totalRunTime := int64(0)

	for _, poolResult := range poolResults {
		if poolResult.Error != nil {
			errs = errors.Join(errs, poolResult.Error)
		} else {
			results = append(results, poolResult.Result)
			totalRunTime += poolResult.RunTime
		}
	}

	return results, totalRunTime, errs
}
