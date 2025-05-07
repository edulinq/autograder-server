package util

import (
	"context"
	"fmt"
	"sync"
)

// The result of running a map function in parallel pool.
// Always call Done() before accessing any results to avoid concurrency issues.
type PoolResult[InputType comparable, OutputType any] struct {
	// The results of the parallel pool.
	Results map[InputType]OutputType

	// A map of work errors using the item's index for item-level errors.
	WorkErrors map[InputType]error

	// Signals that the parallel pool was canceled during execution.
	Canceled bool

	// An internal done channel to signal the result can be accessed.
	done chan any
}

func (this PoolResult[InputType, OutputType]) IsDone() {
	<-this.done
}

// TODO: update comment
// Do a map function (one result for one input) with a parallel pool of workers.
// Unless there is a critical error (the final return value) or cancellation,
// the output will have the same length as and index-match the input, but will have an empty value if there is an error.
// A cancellation will return (partial results, any errors encountered, nil).
// Consult the returned map of errors using the item's index to check for item-level errors.
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
	doneCollectingChan := make(chan bool, poolSize)
	exitWaitGroup := sync.WaitGroup{}

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

	exitWaitGroupChan := make(chan any)
	doneChan := make(chan any)

	output := PoolResult[InputType, OutputType]{
		Results:    make(map[InputType]OutputType, len(workItems)),
		WorkErrors: make(map[InputType]error, 0),
		done:       doneChan,
	}

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
			case <-exitWaitGroupChan:
				// All workers completed in progress work, drain remaining results.
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
	exitWaitGroup.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		go func() {
			defer exitWaitGroup.Done()

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

	// Wait on the wait group in the background so we can select between it and the context.
	// Technically we could wait on the done channel, but this will ensure that all workers have exited.
	go func() {
		exitWaitGroup.Wait()
		close(exitWaitGroupChan)
	}()

	// Wait for either completion or cancellation.
	select {
	case <-ctx.Done():
		// The context was canceled, return partial results.
		output.Canceled = true

		return output, nil
	case <-exitWaitGroupChan:
		// Normal Completion
	}

	return output, nil
}
