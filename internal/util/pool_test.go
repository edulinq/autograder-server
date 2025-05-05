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

	expected := []int{
		1,
		2,
		3,
	}

	workFunc := func(input string) (int, error) {
		return len(input), nil
	}

	for i, testCase := range testCases {
		// Count the number of active threads before running.
		startThreadCount := runtime.NumGoroutine()

		actual, workErrors, err := RunParallelPoolMap(testCase.numThreads, input, context.Background(), workFunc)
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

		if len(workErrors) != 0 {
			test.Errorf("Case %d: Unexpected work errors: '%s'.", i, MustToJSONIndent(workErrors))
			continue
		}

		if !reflect.DeepEqual(expected, actual) {
			test.Errorf("Case %d: Result not as expected. Expected: '%s', Actual: '%s'.", i, MustToJSONIndent(expected), MustToJSONIndent(actual))
		}

		// Check for the thread count last (this gives the workers a small bit of extra time to exit).
		// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
		time.Sleep(25 * time.Millisecond)
		endThreadCount := runtime.NumGoroutine()
		if startThreadCount < endThreadCount {
			test.Errorf("Case %d: Ended with more threads than we started with. Start: %d, End: %d.", i, startThreadCount, endThreadCount)
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
		// Block until the first worker has started.
		workWaitGroup := sync.WaitGroup{}
		workWaitGroup.Add(1)

		workFunc := func(input string) (int, error) {
			// Allow the first input to return a result to test partial results.
			if input == "A" {
				return len(input), nil
			}

			// Sleep to allow the first result to be captured.
			time.Sleep(time.Duration(2) * time.Millisecond)

			// Signal on the second piece of work so that we can make sure the workers have started up before we cancel.
			if input == "BB" {
				workWaitGroup.Done()
			}

			// Sleep for a really long time (for a test).
			time.Sleep(1 * time.Hour)

			return len(input), nil
		}

		ctx, cancelFunc := context.WithCancel(context.Background())

		// Cancel the context as soon as the first worker signals it.
		go func() {
			workWaitGroup.Wait()
			cancelFunc()
		}()

		actual, workErrors, err := RunParallelPoolMap(testCase.numThreads, input, ctx, workFunc)
		if err != nil {
			test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
			continue
		}

		expected := []int{1}

		if !reflect.DeepEqual(actual, expected) {
			test.Errorf("Case %d: Unexpected result. Expected: '%v', actual: '%v'.",
				i, MustToJSONIndent(expected), MustToJSONIndent(actual))
			continue
		}

		expectedWorkErrors := map[int]error{}

		if !reflect.DeepEqual(workErrors, expectedWorkErrors) {
			test.Errorf("Case %d: Unexpected work errors. Expected: '%v', actual: '%v'.",
				i, MustToJSONIndent(expectedWorkErrors), MustToJSONIndent(workErrors))
			continue
		}
	}
}
