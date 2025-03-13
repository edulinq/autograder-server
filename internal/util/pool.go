package util

import (
	"fmt"
	"sync"
)

// Do a map function (one result for one input) with a parallel pool of workers.
// Unless there is a critical error (the final return value),
// the output will have the same length as and index-match the input, but will have an empty value if there is an error.
// Consult the returned map of errors using the item's index to check for item-level errors.
// The underlying collection of input work items must not be modified as this function is running.
func RunParallelPoolMap[InputType any, OutputType any](poolSize int, workItems []InputType, workFunc func(InputType) (OutputType, error)) ([]OutputType, map[int]error, error) {
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
			workQueue <- WorkItem{i, item}
		}
	}()

	// Collect results.
	go func() {
		for i := 0; i < len(workItems); i++ {
			resultItem := <-resultQueue

			results[resultItem.Index] = resultItem.Item
			if resultItem.Error != nil {
				workErrors[resultItem.Index] = resultItem.Error
			}
		}

		// Got all the results, signal completion to all the workers.
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
				case workItem := <-workQueue:
					result, err := workFunc(workItem.Item)
					resultQueue <- ResultItem{workItem.Index, result, err}
				case <-doneChan:
					return
				}
			}
		}()
	}

	// Wait for completion.
	// Technically we could wait on the done channel, but this will ensure that all workers have exited.
	exitWaitGroup.Wait()

	return results, workErrors, nil
}
