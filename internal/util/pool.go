package util

import (
	"context"
	"fmt"
	"sync"
)

// The result of running a map function in parallel pool.
// Always call Done() before accessing any results to avoid concurrency issues.
type PoolResult[OutputType any] struct {
	// The results of the parallel pool.
	Results []OutputType

	// A map of work errors using the item's index for item-level errors.
	WorkErrors map[int]error

	// Signals that the parallel pool was canceled during execution.
	// Use CompletedItems to determine which results were processed.
	Canceled bool

	// A list of indices that indicate which results are completed.
	CompletedItems []bool

	// An internal done function to signal the result can be accessed.
	doneFunc func()
}

func (this PoolResult[OutputType]) Done() {
	if this.doneFunc != nil {
		this.doneFunc()
	}
}

// Do a map function (one result for one input) with a parallel pool of workers.
// Unless there is a critical error (the final return value) or cancellation,
// the output will have the same length as and index-match the input, but will have an empty value if there is an error.
// A cancellation will return (partial results, any errors encountered, nil).
// Consult the returned map of errors using the item's index to check for item-level errors.
// The underlying collection of input work items must not be modified as this function is running.
func RunParallelPoolMap[InputType any, OutputType any](poolSize int, workItems []InputType, ctx context.Context, completeInProgressWork bool, workFunc func(InputType) (OutputType, error)) (PoolResult[OutputType], error) {
	if poolSize <= 0 {
		return PoolResult[OutputType]{}, fmt.Errorf("Pool size must be positive, got %d.", poolSize)
	}

	type WorkItem struct {
		Index int
		Item  InputType
	}

	type ResultItem struct {
		Index int
		Item  OutputType
		Error error
	}

	output := PoolResult[OutputType]{
		Results:        make([]OutputType, len(workItems)),
		WorkErrors:     make(map[int]error),
		CompletedItems: make([]bool, len(workItems)),
		doneFunc:       func() {},
	}

	resultQueue := make(chan ResultItem, poolSize)
	workQueue := make(chan WorkItem, poolSize)
	doneChan := make(chan bool, poolSize)
	exitWaitGroup := sync.WaitGroup{}

	// Load work.
	go func() {
		for i, item := range workItems {
			// Either send a work item, or cancel.
			select {
			case <-ctx.Done():
				return
			case workQueue <- WorkItem{i, item}:
				// Item was already sent on the chan, do nothing here.
			}
		}
	}()

	exitWaitGroupChan := make(chan any)

	// Collect results.
	go func() {
		for i := 0; i < len(workItems); i++ {
			select {
			case <-ctx.Done():
				if !completeInProgressWork {
					return
				}

				// Keep collecting in progress results.
				for {
					select {
					case resultItem := <-resultQueue:
						output.Results[resultItem.Index] = resultItem.Item
						if resultItem.Error != nil {
							output.WorkErrors[resultItem.Index] = resultItem.Error
						}

						output.CompletedItems[resultItem.Index] = true
					case <-exitWaitGroupChan:
						// All workers completed in progress work.
						return
					}
				}
			case resultItem := <-resultQueue:
				output.Results[resultItem.Index] = resultItem.Item
				if resultItem.Error != nil {
					output.WorkErrors[resultItem.Index] = resultItem.Error
				}

				output.CompletedItems[resultItem.Index] = true
			}
		}

		// Got all the results (or canceled), signal completion to all the workers.
		for i := 0; i < (poolSize); i++ {
			doneChan <- true
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
				case workItem := <-workQueue:
					result, err := workFunc(workItem.Item)
					resultQueue <- ResultItem{workItem.Index, result, err}
				case <-doneChan:
					return
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
		if !completeInProgressWork {
			// Return a stable copy with the current progress.
			return PoolResult[OutputType]{
				Results:        output.Results,
				WorkErrors:     output.WorkErrors,
				Canceled:       true,
				CompletedItems: output.CompletedItems,
				doneFunc:       func() {},
			}, nil
		} else {
			output.Canceled = true

			// The done function will block until the workers complete in progress jobs.
			output.doneFunc = func() {
				<-exitWaitGroupChan
			}

			return output, nil
		}
	case <-exitWaitGroupChan:
		// Normal Completion
	}

	return output, nil
}
