package stats

import (
	"reflect"
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
)

func init() {
	simpleBaseMetricsReverse = append([]BaseMetric(nil), simpleBaseMetrics...)
	slices.Reverse(simpleBaseMetricsReverse)
}

func TestApplyBaseQueryBase(test *testing.T) {
	testCases := []struct {
		baseMetrics []BaseMetric
		baseQuery   BaseQuery
		expected    []BaseMetric
	}{
		{
			simpleBaseMetrics,
			BaseQuery{},
			simpleBaseMetrics,
		},

		// Filter

		{
			simpleBaseMetrics,
			BaseQuery{After: timestamp.FromMSecs(200)},
			[]BaseMetric{
				BaseMetric{Timestamp: timestamp.FromMSecs(300)},
				BaseMetric{Timestamp: timestamp.FromMSecs(400)},
				BaseMetric{Timestamp: timestamp.FromMSecs(500)},
			},
		},
		{
			simpleBaseMetrics,
			BaseQuery{Before: timestamp.FromMSecs(300)},
			[]BaseMetric{
				BaseMetric{Timestamp: timestamp.FromMSecs(100)},
				BaseMetric{Timestamp: timestamp.FromMSecs(200)},
			},
		},
		{
			simpleBaseMetrics,
			BaseQuery{
				After:  timestamp.FromMSecs(199),
				Before: timestamp.FromMSecs(301),
			},
			[]BaseMetric{
				BaseMetric{Timestamp: timestamp.FromMSecs(200)},
				BaseMetric{Timestamp: timestamp.FromMSecs(300)},
			},
		},

		// Sort

		{
			simpleBaseMetrics,
			BaseQuery{Sort: -1},
			simpleBaseMetrics,
		},
		{
			simpleBaseMetricsReverse,
			BaseQuery{Sort: 1},
			simpleBaseMetricsReverse,
		},
		{
			simpleBaseMetricsReverse,
			BaseQuery{Sort: 0},
			simpleBaseMetricsReverse,
		},
		{
			simpleBaseMetrics,
			BaseQuery{Sort: 1},
			simpleBaseMetricsReverse,
		},
		{
			simpleBaseMetricsReverse,
			BaseQuery{Sort: -1},
			simpleBaseMetrics,
		},
		{
			simpleBaseMetrics,
			BaseQuery{Sort: 100},
			simpleBaseMetricsReverse,
		},

		// Filter and Sort

		{
			simpleBaseMetrics,
			BaseQuery{
				After:  timestamp.FromMSecs(199),
				Before: timestamp.FromMSecs(301),
				Sort:   1,
			},
			[]BaseMetric{
				BaseMetric{Timestamp: timestamp.FromMSecs(300)},
				BaseMetric{Timestamp: timestamp.FromMSecs(200)},
			},
		},

		// Limit

		{
			simpleBaseMetrics,
			BaseQuery{Limit: 1},
			[]BaseMetric{
				BaseMetric{Timestamp: timestamp.FromMSecs(100)},
			},
		},
		{
			simpleBaseMetrics,
			BaseQuery{
				Limit: 1,
				Sort:  1,
			},
			[]BaseMetric{
				BaseMetric{Timestamp: timestamp.FromMSecs(500)},
			},
		},
	}

	for i, testCase := range testCases {
		actual := ApplyBaseQuery(testCase.baseMetrics, testCase.baseQuery)
		if !reflect.DeepEqual(testCase.expected, actual) {
			test.Errorf("Case %d: Result is not as expected. Expected: '%v', Actual: '%v'.", i, testCase.expected, actual)
			continue
		}
	}
}

var simpleBaseMetrics []BaseMetric = []BaseMetric{
	BaseMetric{Timestamp: timestamp.FromMSecs(100)},
	BaseMetric{Timestamp: timestamp.FromMSecs(200)},
	BaseMetric{Timestamp: timestamp.FromMSecs(300)},
	BaseMetric{Timestamp: timestamp.FromMSecs(400)},
	BaseMetric{Timestamp: timestamp.FromMSecs(500)},
}

var simpleBaseMetricsReverse []BaseMetric = nil
