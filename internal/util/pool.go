package util

import (
	// TEST
	"os"

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

type resultItem[OutputType any] struct {
	Index int
	Item  OutputType
	Error error
}

func (this PoolResult[OutputType]) Done() {
	if this.doneFunc != nil {
		this.doneFunc()
	}
}

func (this PoolResult[OutputType]) GetCompletedResults() []OutputType {
	this.Done()

	if !this.Canceled {
		return this.Results
	}

	completedResults := make([]OutputType, 0, len(this.Results))

	for i := 0; i < len(this.Results) && i < len(this.CompletedItems); i++ {
		if this.CompletedItems[i] {
			completedResults = append(completedResults, this.Results[i])
		}
	}

	return completedResults
}

func (this PoolResult[OutputType]) addResultItem(result resultItem[OutputType]) {
	this.Results[result.Index] = result.Item
	if result.Error != nil {
		this.WorkErrors[result.Index] = result.Error
	}

	this.CompletedItems[result.Index] = true
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

	output := PoolResult[OutputType]{
		Results:        make([]OutputType, len(workItems)),
		WorkErrors:     make(map[int]error),
		CompletedItems: make([]bool, len(workItems)),
		doneFunc:       func() {},
	}

	resultQueue := make(chan resultItem[OutputType], poolSize)
	workQueue := make(chan WorkItem, poolSize)
	doneCollectingChan := make(chan bool, poolSize)
	exitWaitGroup := sync.WaitGroup{}

	// Load work.
	go func() {
		fmt.Fprintf(os.Stderr, "work loader: starting.\n")
		defer fmt.Fprintf(os.Stderr, "work loader: ending.\n")

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
	doneChan := make(chan any)
	go func() {
		fmt.Fprintf(os.Stderr, "result collecter: starting.\n")
		defer fmt.Fprintf(os.Stderr, "result collecter: ending.\n")
		defer func() {
			// Got all the results (or canceled), signal completion to all the workers.
			for i := 0; i < (poolSize); i++ {
				doneCollectingChan <- true
			}

			close(doneChan)
		}()

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
						output.addResultItem(resultItem)
					case <-exitWaitGroupChan:
						// All workers completed in progress work, drain remaining results.
						for {
							select {
							case resultItem := <-resultQueue:
								output.addResultItem(resultItem)
							default:
								return
							}
						}
					}
				}
			case resultItem := <-resultQueue:
				output.addResultItem(resultItem)
			}
		}
	}()

	// Dispatch workers.
	exitWaitGroup.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		go func() {
			fmt.Fprintf(os.Stderr, "worker dispatch %d: starting.\n", i)
			defer fmt.Fprintf(os.Stderr, "worker dispatch %d: ending.\n", i)
			defer exitWaitGroup.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case <-doneCollectingChan:
					return
				case workItem, ok := <-workQueue:
					if !ok {
						return
					}

					result, err := workFunc(workItem.Item)
					// TODO: I believe the resultQueue is full so the worker blocks waiting to put the result in queue.
					// TODO: Need a clean way to signal results are no longer accepted.
					resultQueue <- resultItem[OutputType]{workItem.Index, result, err}
				}
			}
		}()
	}

	// Wait on the wait group in the background so we can select between it and the context.
	// Technically we could wait on the done channel, but this will ensure that all workers have exited.
	go func() {
		fmt.Fprintf(os.Stderr, "work waiter: starting.\n")
		defer fmt.Fprintf(os.Stderr, "work waiter: ending.\n")
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
				<-doneChan
			}

			return output, nil
		}
	case <-exitWaitGroupChan:
		// Normal Completion
	}

	return output, nil
}
