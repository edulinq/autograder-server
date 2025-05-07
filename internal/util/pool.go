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

	// An internal done channel to signal the result can be accessed.
	done chan any
}

func (this PoolResult[InputType, OutputType]) IsDone() {
	<-this.done
}

// Do a map function (one result for one input) with a parallel pool of workers.
// Unless there is a critical error (the final return value) or cancellation,
// every input will either be in Results or WorkErrors.
// A cancellation will stop new work from starting but complete in progress work asynchronously.
// The partial results from a cancellation are available after IsDone() returns.
// Consult the returned map of errors using the item to check for item-level errors.
// The underlying collection of input work items must not be modified as this function is running.
func RunParallelPoolMap[InputType comparable, OutputType any](poolSize int, workItems []InputType, ctx context.Context, workFunc func(InputType) (OutputType, error)) (PoolResult[InputType, OutputType], error) {
	doneChan := make(chan any)

	output := PoolResult[InputType, OutputType]{
		Results:    make(map[InputType]OutputType, len(workItems)),
		WorkErrors: make(map[InputType]error, 0),
		done:       doneChan,
	}

	if poolSize <= 0 {
		return output, fmt.Errorf("Pool size must be positive, got %d.", poolSize)
	}

	type ResultItem struct {
		Input  InputType
		Result OutputType
		Error  error
	}

	resultQueue := make(chan ResultItem, poolSize)
	workQueue := make(chan InputType, poolSize)
	doneCollectingChan := make(chan bool, poolSize)
	workerExitWaitGroup := sync.WaitGroup{}

	// Load work.
	go func() {
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

	workerExitWaitGroupChan := make(chan any)

	// Collect results.
	go func() {
		defer func() {
			// Got all the results (or canceled), signal completion to all the workers.
			for i := 0; i < (poolSize); i++ {
				doneCollectingChan <- true
			}

			close(doneChan)
		}()

		for i := 0; i < len(workItems); i++ {
			select {
			case <-workerExitWaitGroupChan:
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
	for i := 0; i < poolSize; i++ {
		go func() {
			defer workerExitWaitGroup.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case <-doneCollectingChan:
					return
				case workItem := <-workQueue:
					result, err := workFunc(workItem)
					resultQueue <- ResultItem{workItem, result, err}
				}
			}
		}()
	}

	// Wait on the done chan in the background so we can select between it and the context.
	// Technically we could wait on the worker wait group, but this will ensure that all results are collected.
	go func() {
		workerExitWaitGroup.Wait()
		close(workerExitWaitGroupChan)
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

	return output, nil
}
