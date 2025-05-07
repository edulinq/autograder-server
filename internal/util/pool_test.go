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
		time.Sleep(25 * time.Millisecond)
		endThreadCount := runtime.NumGoroutine()
		if startThreadCount < endThreadCount {
			test.Errorf("Case %d: Ended with more threads than we started with. Start: %d, End: %d.",
				i, startThreadCount, endThreadCount)
			continue
		}
	}
}

func TestRunParallelPoolMapCancel(test *testing.T) {
	testCases := []struct {
		numThreads int
	}{
		{1},
		{2},
		{3},
		{4},
		{10},
	}

	input := []string{
		"A",
		"BB",
		"CCC",
	}

	for i, testCase := range testCases {
		// Count the number of active threads before running.
		startThreadCount := runtime.NumGoroutine()

		// A channel to know which jobs started.
		startedChan := make(chan string, len(input))

		// Block until the first worker has started.
		workWaitGroup := sync.WaitGroup{}
		workWaitGroup.Add(1)

		workFunc := func(input string) (int, error) {
			startedChan <- input

			// Signal on the first piece of work so that we can make sure the workers have started up before we cancel.
			if input == "A" {
				workWaitGroup.Done()
				time.Sleep(time.Duration(5) * time.Millisecond)
				return len(input), nil
			}

			time.Sleep(time.Duration(5) * time.Millisecond)

			return len(input), nil
		}

		ctx, cancelFunc := context.WithCancel(context.Background())

		// Cancel the context as soon as the first worker signals it.
		go func() {
			workWaitGroup.Wait()
			cancelFunc()
		}()

		output, err := RunParallelPoolMap(testCase.numThreads, input, ctx, workFunc)
		if err != nil {
			test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
			continue
		}

		output.IsDone()

		close(startedChan)

		// Workers race to start before the cancellation.
		// Ensure all started work is completed.
		expectedResults := map[string]int{}
		for input := range startedChan {
			expectedResults[input] = len(input)
		}

		expected := PoolResult[string, int]{
			Results:    expectedResults,
			WorkErrors: map[string]error{},
			Canceled:   true,
		}

		// Clear output done channel for comparison check.
		output.done = nil

		if !reflect.DeepEqual(output, expected) {
			test.Errorf("Case %d: Unexpected results. Expected: '%v', actual: '%v'.",
				i, MustToJSONIndent(expected), MustToJSONIndent(output))
			continue
		}

		// Check for the thread count last (this gives the workers a small bit of extra time to exit).
		// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
		endThreadCount := runtime.NumGoroutine()
		if startThreadCount < endThreadCount {
			test.Errorf("Case %d: Ended with more threads than we started with. Start: %d, End: %d.",
				i, startThreadCount, endThreadCount)
			continue
		}
	}
}
