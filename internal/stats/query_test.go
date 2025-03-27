package stats

import (
	"reflect"
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
)

func init() {
	simpleBaseMetricsReverse = append([]Metric(nil), simpleBaseMetrics...)
	slices.Reverse(simpleBaseMetricsReverse)
}

func TestApplyBaseQueryBase(test *testing.T) {
	testCases := []struct {
		metrics  []Metric
		Query    Query
		expected []Metric
	}{
		{
			simpleBaseMetrics,
			Query{},
			simpleBaseMetrics,
		},

		// Filter
		{
			simpleBaseMetrics,
			Query{After: timestamp.FromMSecs(200)},
			[]Metric{
				Metric{Timestamp: timestamp.FromMSecs(300)},
				Metric{Timestamp: timestamp.FromMSecs(400)},
				Metric{Timestamp: timestamp.FromMSecs(500)},
			},
		},
		{
			simpleBaseMetrics,
			Query{Before: timestamp.FromMSecs(300)},
			[]Metric{
				Metric{Timestamp: timestamp.FromMSecs(100)},
				Metric{Timestamp: timestamp.FromMSecs(200)},
			},
		},
		{
			simpleBaseMetrics,
			Query{
				After:  timestamp.FromMSecs(199),
				Before: timestamp.FromMSecs(301),
			},
			[]Metric{
				Metric{Timestamp: timestamp.FromMSecs(200)},
				Metric{Timestamp: timestamp.FromMSecs(300)},
			},
		},

		// Sort
		{
			simpleBaseMetrics,
			Query{Sort: -1},
			simpleBaseMetrics,
		},
		{
			simpleBaseMetricsReverse,
			Query{Sort: 1},
			simpleBaseMetricsReverse,
		},
		{
			simpleBaseMetricsReverse,
			Query{Sort: 0},
			simpleBaseMetricsReverse,
		},
		{
			simpleBaseMetrics,
			Query{Sort: 1},
			simpleBaseMetricsReverse,
		},
		{
			simpleBaseMetricsReverse,
			Query{Sort: -1},
			simpleBaseMetrics,
		},
		{
			simpleBaseMetrics,
			Query{Sort: 100},
			simpleBaseMetricsReverse,
		},

		// Filter and Sort
		{
			simpleBaseMetrics,
			Query{
				After:  timestamp.FromMSecs(199),
				Before: timestamp.FromMSecs(301),
				Sort:   1,
			},
			[]Metric{
				Metric{Timestamp: timestamp.FromMSecs(300)},
				Metric{Timestamp: timestamp.FromMSecs(200)},
			},
		},

		// Limit
		{
			simpleBaseMetrics,
			Query{Limit: 1},
			[]Metric{
				Metric{Timestamp: timestamp.FromMSecs(100)},
			},
		},
		{
			simpleBaseMetrics,
			Query{
				Limit: 1,
				Sort:  1,
			},
			[]Metric{
				Metric{Timestamp: timestamp.FromMSecs(500)},
			},
		},
	}

	for i, testCase := range testCases {
		actual := ApplyBaseQuery(testCase.metrics, testCase.Query)
		if !reflect.DeepEqual(testCase.expected, actual) {
			test.Errorf("Case %d: Result is not as expected. Expected: '%v', Actual: '%v'.", i, testCase.expected, actual)
			continue
		}
	}
}

var simpleBaseMetrics []Metric = []Metric{
	Metric{Timestamp: timestamp.FromMSecs(100)},
	Metric{Timestamp: timestamp.FromMSecs(200)},
	Metric{Timestamp: timestamp.FromMSecs(300)},
	Metric{Timestamp: timestamp.FromMSecs(400)},
	Metric{Timestamp: timestamp.FromMSecs(500)},
}

var simpleBaseMetricsReverse []Metric = nil
