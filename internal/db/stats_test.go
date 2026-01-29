package db

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestStoreMetricTypeSystemCPU(test *testing.T) {
	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.MetricTypeSystemCPU,
	}

	runStoreStatsTests(test, &metric)
}

func (this *DBTests) DBTestStoreMetricTypeSystemMemory(test *testing.T) {
	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.MetricTypeSystemMemory,
	}

	runStoreStatsTests(test, &metric)
}

func (this *DBTests) DBTestStoreMetricTypeSystemNetworkIn(test *testing.T) {
	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.MetricTypeSystemNetworkIn,
	}

	runStoreStatsTests(test, &metric)
}

func (this *DBTests) DBTestStoreMetricTypeSystemNetworkOut(test *testing.T) {
	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.MetricTypeSystemNetworkOut,
	}

	runStoreStatsTests(test, &metric)
}

func (this *DBTests) DBTestStoreMetricTypeAPIRequest(test *testing.T) {
	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.MetricTypeAPIRequest,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeSender:       "1",
			stats.MetricAttributeEndpoint:     "E1",
			stats.MetricAttributeUserEmail:    "U",
			stats.MetricAttributeAssignmentID: "A",
			stats.MetricAttributeCourseID:     "C",
			stats.MetricAttributeLocator:      "11",
		},
	}

	runStoreStatsTests(test, &metric)
}

func (this *DBTests) DBTestStoreMetricTypeTaskTime(test *testing.T) {
	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeUserEmail:    "U",
			stats.MetricAttributeAssignmentID: "A",
			stats.MetricAttributeCourseID:     "C",
		},
	}

	runStoreStatsTests(test, &metric)
}

func (this *DBTests) DBTestStoreMetricTypeGradingTime(test *testing.T) {
	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.MetricTypeGradingTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeUserEmail:    "U",
			stats.MetricAttributeAssignmentID: "A",
			stats.MetricAttributeCourseID:     "C",
		},
	}

	runStoreStatsTests(test, &metric)
}

func (this *DBTests) DBTestStoreMetricTypeCodeAnalysisTime(test *testing.T) {
	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.MetricTypeCodeAnalysisTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeUserEmail:    "U",
			stats.MetricAttributeAssignmentID: "A",
			stats.MetricAttributeCourseID:     "C",
		},
	}

	runStoreStatsTests(test, &metric)
}

func (this *DBTests) DBTestGetMetricFailure(test *testing.T) {
	Clear()
	defer Clear()

	query := stats.Query{}

	_, err := GetMetrics(query)
	if err == nil {
		test.Fatalf("Expected error due to missing query type, but got none.")
	}

	expectedSubstring := "No metric type was given."
	if !strings.Contains(err.Error(), expectedSubstring) {
		test.Errorf("Did not get expected error substring. Expected: '%s', Actual: '%s'.", expectedSubstring, err.Error())
	}
}

func (this *DBTests) DBTestStoreMetricFailure(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Value:     100,
	}

	err := StoreMetric(&metric)
	if err == nil {
		test.Fatalf("Expected error due to missing metric type, but got none.")
	}

	expectedSubstring := "No metric type was given."
	if !strings.Contains(err.Error(), expectedSubstring) {
		test.Errorf("Expected error to contain '%s', got: '%s'", expectedSubstring, err.Error())
	}
}

func DBTestAsyncStoreMetric(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.MetricTypeAPIRequest,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeSender:       "1",
			stats.MetricAttributeEndpoint:     "E1",
			stats.MetricAttributeUserEmail:    "U",
			stats.MetricAttributeAssignmentID: "A",
			stats.MetricAttributeCourseID:     "C",
			stats.MetricAttributeLocator:      "11",
		},
	}

	query := stats.Query{
		Type: stats.MetricTypeAPIRequest,
	}

	stats.AsyncStoreMetric(&metric)

	records, err := GetMetrics(query)
	if err != nil {
		test.Fatalf("Failed to fetch stats: '%v'.", err)
	}

	if len(records) != 1 {
		test.Fatalf("Did not get the correct number of records. Expected: 1, Actual: %d.", len(records))
	}

	if !reflect.DeepEqual(records[0], metric) {
		test.Fatalf("Did not get the expected record back. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(metric), util.MustToJSONIndent(*records[0]))
	}
}

func runStoreStatsTests(test *testing.T, metric *stats.Metric) {
	Clear()
	defer Clear()

	err := StoreMetric(metric)
	if err != nil {
		test.Fatalf("Failed to store stats: '%v'.", err)
	}

	query := stats.Query{
		Type: metric.Type,
	}

	records, err := GetMetrics(query)
	if err != nil {
		test.Fatalf("Failed to fetch stats: '%v'.", err)
	}

	if len(records) != 1 {
		test.Fatalf("Did not get the correct number of records. Expected: 1, Actual: %d.", len(records))
	}

	if !reflect.DeepEqual(records[0], metric) {
		test.Fatalf("Did not get the expected record back. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(metric), util.MustToJSONIndent(*records[0]))
	}
}

func (this *DBTests) DBTestGetTestingMetrics(test *testing.T) {
	testCases := []struct {
		query    stats.Query
		expected []*stats.Metric
	}{
		{
			stats.Query{
				Type: stats.MetricTypeAPIRequest,
			},
			[]*stats.Metric{TESTING_STATS_METRICS[0]},
		},
		{
			stats.Query{
				Type: stats.MetricTypeSystemCPU,
			},
			[]*stats.Metric{
				TESTING_STATS_METRICS[7],
				TESTING_STATS_METRICS[11],
			},
		},
	}

	for i, testCase := range testCases {
		actual, err := GetMetricsFull(testCase.query, true)
		if err != nil {
			test.Errorf("Case %d: Failed to get metrics: '%v'.", i, err)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, actual) {
			test.Errorf("Case %d: Unexpected metrics. Expected: %s, Actual: %s.", i,
				util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(actual))
			continue
		}
	}
}
