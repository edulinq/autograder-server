package util

import (
	"context"
	"fmt"
	"sync"
)

// The result of running a map function in parallel pool.
// Always call IsDone() before accessing any results to avoid concurrency issues.
type PoolResult[InputType comparable, OutputType any] struct {
	// The results of the parallel pool.
	Results map[InputType]OutputType

	// A map of work errors using the item for item-level errors.
	WorkErrors map[InputType]error

	// Signals that the parallel pool was canceled during execution.
	Canceled bool

	// An internal done channel to signal the results can be accessed.
	done chan any
}

// Use this method to block until the results can be accessed.
func (this PoolResult[InputType, OutputType]) IsDone() {
	<-this.done
}

// Do a map function (one result for one input) with a parallel pool of workers.
// Unless there is a critical error (the final return value) or cancellation,
// every input will either be in PoolResult.Results or PoolResult.WorkErrors.
// A cancellation will stop new work from starting but complete in progress work asynchronously.
// The partial results from a cancellation are available after PoolResult.IsDone() returns.
// Consult the returned map of errors using the item to check for item-level errors.
// Work level errors will not generate a critical error, so both must be checked independently.
// The underlying collection of input work items must not be modified as this function is running.
func RunParallelPoolMap[InputType comparable, OutputType any](poolSize int, workItems []InputType, ctx context.Context, workFunc func(InputType) (OutputType, error)) (*PoolResult[InputType, OutputType], error) {
	if poolSize <= 0 {
		return nil, fmt.Errorf("Pool size must be positive, got %d.", poolSize)
	}

	type ResultItem struct {
		Input  InputType
		Result OutputType
		Error  error
	}

	resultQueue := make(chan ResultItem, poolSize)
	workQueue := make(chan InputType, poolSize)

	// The done working channel is used by the pool workers to signal work is complete to the result collectors.
	doneWorkingChan := make(chan any)

	// The done channel is used to signal that the all results are collected and accessible.
	// It also waits to signal until all worker threads join the main thread.
	doneChan := make(chan any)

	// A wait group used to determine when the workers have completed their work.
	// Note that completion can be natural or due to a cancellation.
	workerExitWaitGroup := sync.WaitGroup{}
	workerExitWaitGroup.Add(poolSize)

	// There will always be poolSize + 2 threads started.
	// A thread will be started for the work loader, the result collector, and each worker in the pool.
	// Every one of these work threads must call threadExitWaitGroup.Done() to avoid a deadlock.
	threadExitWaitGroup := sync.WaitGroup{}
	threadExitWaitGroup.Add(poolSize + 2)

	output := PoolResult[InputType, OutputType]{
		Results:    make(map[InputType]OutputType, len(workItems)),
		WorkErrors: make(map[InputType]error, 0),
		done:       doneChan,
	}

	// Load work.
	go func() {
		defer threadExitWaitGroup.Done()
		// Signal to the workers that there is no more work.
		defer close(workQueue)

		for _, item := range workItems {
			// Either send a work item, or cancel.
			select {
			case <-ctx.Done():
				return
			case workQueue <- item:
				// Item was already sent on the chan, do nothing here.
			}
		}
	}()

	// Collect results.
	go func() {
		defer func() {
			defer threadExitWaitGroup.Done()

			// All workers completed, drain remaining results.
			// If the context was canceled, the number of results may be less than the number of work items.
			for {
				select {
				case resultItem := <-resultQueue:
					if resultItem.Error != nil {
						output.WorkErrors[resultItem.Input] = resultItem.Error
					} else {
						output.Results[resultItem.Input] = resultItem.Result
					}
				default:
					return
				}
			}
		}()

		for i := 0; i < len(workItems); i++ {
			select {
			case <-doneWorkingChan:
				return
			case resultItem := <-resultQueue:
				if resultItem.Error != nil {
					output.WorkErrors[resultItem.Input] = resultItem.Error
				} else {
					output.Results[resultItem.Input] = resultItem.Result
				}
			}
		}
	}()

	// Dispatch workers.
	for i := 0; i < poolSize; i++ {
		go func() {
			defer threadExitWaitGroup.Done()
			defer workerExitWaitGroup.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case workItem, ok := <-workQueue:
					// The work queue was closed which signals there is no more work.
					if !ok {
						return
					}

					// Pulling work from the queue can be selected over a context cancellation.
					// Check the context for cancellations before starting new work.
					if ctx.Err() != nil {
						return
					}

					result, err := workFunc(workItem)
					resultQueue <- ResultItem{workItem, result, err}
				}
			}
		}()
	}

	// Wait on the wait group in the background so we can select between it and the context.
	// Technically we could wait on the done channel, but this will ensure that all workers have exited.
	go func() {
		// Once all pool workers have completed their work, signal to collect the remaining results.
		workerExitWaitGroup.Wait()
		close(doneWorkingChan)

		// Wait for all worker threads to signal they have exited.
		threadExitWaitGroup.Wait()

		// Close the done channel to signal that the output is accessible and worker threads are joined.
		close(doneChan)
	}()

	// Wait for either completion or cancellation.
	select {
	case <-ctx.Done():
		// The context was canceled, return partial results.
		output.Canceled = true

		return &output, nil
	case <-doneChan:
		// Normal Completion
	}

	return &output, nil
}
