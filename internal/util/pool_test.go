package util

import (
	"reflect"
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

	expected := []int{
		1,
		2,
		3,
	}

	workFunc := func(input string) (int, error) {
		return len(input), nil
	}

	for i, testCase := range testCases {
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
	}
}
