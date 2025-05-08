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
	if this.done != nil {
		<-this.done
	}
}

// Do a map function (one result for one input) with a parallel pool of workers.
// Unless there is a critical error (the final return value) or cancellation,
// every input will be in PoolResult.Results.
// A cancellation will stop new work from starting but complete in progress work asynchronously.
// The partial results from a cancellation are available after PoolResult.IsDone() returns.
// Consult the returned map of errors using the item to check for item-level errors.
// Work level errors will not generate a non-nil return error, so both must be checked independently.
// The underlying collection of input work items must not be modified as this function is running.
func RunParallelPoolMap[InputType comparable, OutputType any](poolSize int, workItems []InputType, ctx context.Context, workFunc func(InputType) (OutputType, error)) (PoolResult[InputType, OutputType], error) {
	if poolSize <= 0 {
		return PoolResult[InputType, OutputType]{}, fmt.Errorf("Pool size must be positive, got %d.", poolSize)
	}

	type ResultItem struct {
		Input  InputType
		Result OutputType
		Error  error
	}

	resultQueue := make(chan ResultItem, poolSize)
	workQueue := make(chan InputType, poolSize)
	doneChan := make(chan any)
	doneWorkingChan := make(chan any)

	workerExitWaitGroup := sync.WaitGroup{}
	threadExitWaitGroup := sync.WaitGroup{}

	output := PoolResult[InputType, OutputType]{
		Results:    make(map[InputType]OutputType, len(workItems)),
		WorkErrors: make(map[InputType]error, 0),
		done:       doneChan,
	}

	// Load work.
	threadExitWaitGroup.Add(1)
	go func() {
		defer threadExitWaitGroup.Done()

		for _, item := range workItems {
			// Either send a work item, or cancel.
			select {
			case <-ctx.Done():
				return
			case workQueue <- item:
				// Item was already sent on the chan, do nothing here.
			}
		}

		// Signal to the workers that there is no more work.
		close(workQueue)
	}()

	// Collect results.
	threadExitWaitGroup.Add(1)
	go func() {
		defer func() {
			// All workers completed, drain remaining results.
			// If the context was canceled, the number of results may be less than the number of work items.
			for {
				select {
				case resultItem := <-resultQueue:
					output.Results[resultItem.Input] = resultItem.Result
					if resultItem.Error != nil {
						output.WorkErrors[resultItem.Input] = resultItem.Error
					}
				default:
					return
				}
			}
		}()

		// Close the done channel to signal that the output is accessible.
		defer close(doneChan)
		defer threadExitWaitGroup.Done()

		for i := 0; i < len(workItems); i++ {
			select {
			case <-doneWorkingChan:
				return
			case resultItem := <-resultQueue:
				output.Results[resultItem.Input] = resultItem.Result
				if resultItem.Error != nil {
					output.WorkErrors[resultItem.Input] = resultItem.Error
				}
			}
		}
	}()

	// Dispatch workers.
	workerExitWaitGroup.Add(poolSize)
	threadExitWaitGroup.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		go func() {
			defer workerExitWaitGroup.Done()
			defer threadExitWaitGroup.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case workItem, ok := <-workQueue:
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
	threadExitWaitGroup.Add(1)
	go func() {
		defer threadExitWaitGroup.Done()

		workerExitWaitGroup.Wait()
		close(doneWorkingChan)
	}()

	// Wait for either completion or cancellation.
	select {
	case <-ctx.Done():
		// The context was canceled, return partial results.
		output.Canceled = true

		return output, nil
	case <-doneChan:
		// Normal Completion
	}

	threadExitWaitGroup.Wait()

	return output, nil
}
