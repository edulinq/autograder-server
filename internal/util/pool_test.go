package util

import (
	"context"
	"reflect"
	"runtime"
	"testing"
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

	expected := &PoolResult[string, int]{
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

	// Count the number of active threads before running any tests.
	overallStartThreadCount := runtime.NumGoroutine()

	for i, testCase := range testCases {
		// Count the number of active threads before running the test case.
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
		output.Done = nil

		if !reflect.DeepEqual(expected, output) {
			test.Errorf("Case %d: Result not as expected. Expected: '%s', actual: '%s'.",
				i, MustToJSONIndent(expected), MustToJSONIndent(output))
			continue
		}

		// Check for the thread count last (this gives the workers a small bit of extra time to exit).
		// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
		// The final thread that signals the output.IsDone() may not have enough time to exit,
		// so we are allowed to end with one more thread than when we started.
		endThreadCount := runtime.NumGoroutine()
		allowedEndThreadCount := endThreadCount - 1
		if startThreadCount < allowedEndThreadCount {
			test.Errorf("Case %d: Ended with more threads than we started with. Start: %d, End: %d.",
				i, startThreadCount, allowedEndThreadCount)
			continue
		}
	}

	// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
	// The final thread that signals the output.IsDone() may not have enough time to exit,
	// so we are allowed to end with one more thread than when we started.
	overallEndThreadCount := runtime.NumGoroutine()
	allowedOverallEndThreadCount := overallEndThreadCount - 1
	if overallStartThreadCount < allowedOverallEndThreadCount {
		test.Fatalf("Ended with more threads than we started with. Start: %d, End: %d.",
			overallStartThreadCount, allowedOverallEndThreadCount)
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

	ctx, cancelFunc := context.WithCancel(context.Background())

	// Work will not start on input 'CCC'.
	workFunc := func(input string) (int, error) {
		// Cancel on the second piece of work so that work on input 'A' will complete normally.
		if input == "BB" {
			cancelFunc()
		}

		return len(input), nil
	}

	output, err := RunParallelPoolMap(1, input, ctx, workFunc)
	if err != nil {
		test.Fatalf("Got an unexpected error: '%v'.", err)
	}

	output.IsDone()

	expected := &PoolResult[string, int]{
		Results: map[string]int{
			"A":  1,
			"BB": 2,
		},
		WorkErrors: map[string]error{},
		Canceled:   true,
	}

	// Clear output done channel for comparison check.
	output.Done = nil

	if !reflect.DeepEqual(output, expected) {
		test.Fatalf("Unexpected results. Expected: '%v', actual: '%v'.",
			MustToJSONIndent(expected), MustToJSONIndent(output))
	}

	// Check for the thread count last (this gives the workers a small bit of extra time to exit).
	// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
	// The final thread that signals the output.IsDone() may not have enough time to exit,
	// so we are allowed to end with one more thread than when we started.
	endThreadCount := runtime.NumGoroutine()
	allowedEndThreadCount := endThreadCount - 1
	if startThreadCount < allowedEndThreadCount {
		test.Fatalf("Ended with more threads than we started with. Start: %d, End: %d.",
			startThreadCount, allowedEndThreadCount)
	}
}
