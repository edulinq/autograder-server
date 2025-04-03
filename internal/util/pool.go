package util

import (
	"context"
	"fmt"
	"sync"
)

// Do a map function (one result for one input) with a parallel pool of workers.
// Unless there is a critical error (the final return value) or cancellation,
// the output will have the same length as and index-match the input, but will have an empty value if there is an error.
// A cancellation will return (nil, nil, nil).
// Consult the returned map of errors using the item's index to check for item-level errors.
// The underlying collection of input work items must not be modified as this function is running.
func RunParallelPoolMap[InputType any, OutputType any](poolSize int, workItems []InputType, ctx context.Context, workFunc func(InputType) (OutputType, error)) ([]OutputType, map[int]error, error) {
	if poolSize <= 0 {
		return nil, nil, fmt.Errorf("Pool size must be positive, got %d.", poolSize)
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

	results := make([]OutputType, len(workItems))
	workErrors := make(map[int]error)

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

	// Collect results.
	go func() {
		for i := 0; i < len(workItems); i++ {
			select {
			case <-ctx.Done():
				return
			case resultItem := <-resultQueue:
				results[resultItem.Index] = resultItem.Item
				if resultItem.Error != nil {
					workErrors[resultItem.Index] = resultItem.Error
				}
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
	exitWaitGroupChan := make(chan bool, 1)
	go func() {
		exitWaitGroup.Wait()
		exitWaitGroupChan <- true
	}()

	// Wait for either completion or cancellation.
	select {
	case <-ctx.Done():
		// The context was canceled, return nothing.
		return nil, nil, nil
	case <-exitWaitGroupChan:
		// Normal Completion
	}

	return results, workErrors, nil
}
