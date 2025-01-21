package util

import (
	"testing"
)

func TestComputeAggregatesBase(test *testing.T) {
	testCases := []struct {
		input    []float64
		expected AggregateValues
	}{
		{
			[]float64{},
			AggregateValues{},
		},
		{
			nil,
			AggregateValues{},
		},
		{
			[]float64{1.0},
			AggregateValues{
				Count:  1,
				Mean:   1.0,
				Median: 1.0,
				Min:    1.0,
				Max:    1.0,
			},
		},
		{
			[]float64{1.0, 2.0},
			AggregateValues{
				Count:  2,
				Mean:   1.5,
				Median: 1.5,
				Min:    1.0,
				Max:    2.0,
			},
		},
		{
			[]float64{3.0, 1.0, 2.0},
			AggregateValues{
				Count:  3,
				Mean:   2.0,
				Median: 2.0,
				Min:    1.0,
				Max:    3.0,
			},
		},
	}

	for i, testCase := range testCases {
		actual := ComputeAggregates(testCase.input)
		if !testCase.expected.Equals(actual) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', Actual: '%s'.",
				i, MustToJSONIndent(testCase.expected), MustToJSONIndent(actual))
		}
	}
}
