package util

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestRunParallelPoolMapBase(test *testing.T) {
	testCases := []struct {
		numThreads int
		hasError   bool
	}{
		{-1, true},
		{0, true},
		{1, false},
		{2, false},
		{3, false},
		{4, false},
		{10, false},
	}

	input := []string{
		"A",
		"BB",
		"CCC",
	}

	expected := PoolResult[string, int]{
		Results: map[string]int{
			"A":   1,
			"BB":  2,
			"CCC": 3,
		},
		WorkErrors: map[string]error{},
		Canceled:   false,
	}

	workFunc := func(input string) (int, error) {
		return len(input), nil
	}

	for i, testCase := range testCases {
		// Count the number of active threads before running.
		startThreadCount := runtime.NumGoroutine()

		output, err := RunParallelPoolMap(testCase.numThreads, input, context.Background(), workFunc)
		if err != nil {
			if !testCase.hasError {
				test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
			}

			continue
		}

		if testCase.hasError {
			test.Errorf("Case %d: Did not get an expected error.", i)
			continue
		}

		output.IsDone()

		// Clear output done channel for comparison check.
		output.done = nil

		if !reflect.DeepEqual(expected, output) {
			test.Errorf("Case %d: Result not as expected. Expected: '%s', actual: '%s'.",
				i, MustToJSONIndent(expected), MustToJSONIndent(output))
			continue
		}

		// Check for the thread count last (this gives the workers a small bit of extra time to exit).
		// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
		time.Sleep(25 * time.Nanosecond)
		endThreadCount := runtime.NumGoroutine()
		if startThreadCount < endThreadCount {
			test.Errorf("Case %d: Ended with more threads than we started with. Start: %d, End: %d.",
				i, startThreadCount, endThreadCount)
			continue
		}
	}
}

func TestRunParallelPoolMapCancel(test *testing.T) {
	input := []string{
		"A",
		"BB",
		"CCC",
	}

	// Count the number of active threads before running.
	startThreadCount := runtime.NumGoroutine()

	// Block until the first worker has started.
	workWaitGroup := sync.WaitGroup{}
	workWaitGroup.Add(1)

	// Block until the cancellation is called.
	cancelWaitGroup := sync.WaitGroup{}
	cancelWaitGroup.Add(1)

	// Input 'A' will complete normally and 'BB' will be completed despite the cancellation signal during execution.
	// Work will not start on input 'CCC'.
	workFunc := func(input string) (int, error) {
		// Signal on the second piece of work so that we can make sure the workers have started up before we cancel.
		if input == "BB" {
			workWaitGroup.Done()
			// Wait until the cancellation is triggered.
			cancelWaitGroup.Wait()
		}

		return len(input), nil
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	// Cancel the context as soon as the first worker signals it.
	go func() {
		workWaitGroup.Wait()
		cancelFunc()
		// Signal to continue working after the cancellation signal goes through.
		cancelWaitGroup.Done()
	}()

	output, err := RunParallelPoolMap(1, input, ctx, workFunc)
	if err != nil {
		test.Fatalf("Got an unexpected error: '%v'.", err)
	}

	output.IsDone()

	expected := PoolResult[string, int]{
		Results: map[string]int{
			"A":  1,
			"BB": 2,
		},
		WorkErrors: map[string]error{},
		Canceled:   true,
	}

	// Clear output done channel for comparison check.
	output.done = nil

	if !reflect.DeepEqual(output, expected) {
		test.Fatalf("Unexpected results. Expected: '%v', actual: '%v'.",
			MustToJSONIndent(expected), MustToJSONIndent(output))
	}

	// Check for the thread count last (this gives the workers a small bit of extra time to exit).
	// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
	time.Sleep(25 * time.Nanosecond)
	endThreadCount := runtime.NumGoroutine()
	if startThreadCount < endThreadCount {
		test.Fatalf("Ended with more threads than we started with. Start: %d, End: %d.",
			startThreadCount, endThreadCount)
	}
}
