package stats

import (
	"reflect"
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
)

func init() {
	simpleMetricsReverse = append([]Metric(nil), simpleMetrics...)
	slices.Reverse(simpleMetricsReverse)
}

func TestLimitAndSort(test *testing.T) {
	testCases := []struct {
		metrics  []Metric
		Query    Query
		expected []Metric
	}{
		{
			simpleMetrics,
			Query{},
			simpleMetrics,
		},

		// Filter
		{
			simpleMetrics,
			Query{After: timestamp.FromMSecs(200)},
			[]Metric{
				Metric{Timestamp: timestamp.FromMSecs(300)},
				Metric{Timestamp: timestamp.FromMSecs(400)},
				Metric{Timestamp: timestamp.FromMSecs(500)},
			},
		},
		{
			simpleMetrics,
			Query{Before: timestamp.FromMSecs(300)},
			[]Metric{
				Metric{Timestamp: timestamp.FromMSecs(100)},
				Metric{Timestamp: timestamp.FromMSecs(200)},
			},
		},
		{
			simpleMetrics,
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
			simpleMetrics,
			Query{Sort: -1},
			simpleMetrics,
		},
		{
			simpleMetricsReverse,
			Query{Sort: 1},
			simpleMetricsReverse,
		},
		{
			simpleMetricsReverse,
			Query{Sort: 0},
			simpleMetricsReverse,
		},
		{
			simpleMetrics,
			Query{Sort: 1},
			simpleMetricsReverse,
		},
		{
			simpleMetricsReverse,
			Query{Sort: -1},
			simpleMetrics,
		},
		{
			simpleMetrics,
			Query{Sort: 100},
			simpleMetricsReverse,
		},

		// Filter and Sort
		{
			simpleMetrics,
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
			simpleMetrics,
			Query{Limit: 1},
			[]Metric{
				Metric{Timestamp: timestamp.FromMSecs(100)},
			},
		},
		{
			simpleMetrics,
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
		actual := LimitAndSort(testCase.metrics, testCase.Query)
		if !reflect.DeepEqual(testCase.expected, actual) {
			test.Errorf("Case %d: Result is not as expected. Expected: '%v', Actual: '%v'.", i, testCase.expected, actual)
			continue
		}
	}
}

var simpleMetrics []Metric = []Metric{
	Metric{Timestamp: timestamp.FromMSecs(100)},
	Metric{Timestamp: timestamp.FromMSecs(200)},
	Metric{Timestamp: timestamp.FromMSecs(300)},
	Metric{Timestamp: timestamp.FromMSecs(400)},
	Metric{Timestamp: timestamp.FromMSecs(500)},
}

var simpleMetricsReverse []Metric = nil
