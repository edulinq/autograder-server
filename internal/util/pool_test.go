package util

import (
	"reflect"
	"runtime"
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

		actual, workErrors, err := RunParallelPoolMap(testCase.numThreads, input, workFunc)
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
