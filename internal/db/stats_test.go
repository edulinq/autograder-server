package db

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestStoreSystemStats(test *testing.T) {
	Clear()
	defer Clear()

	testRecord := stats.SystemMetrics{
		Metric: stats.Metric{
			Timestamp: timestamp.Now(),
		},
		CPUPercent:       1,
		MemPercent:       2,
		NetBytesSent:     3,
		NetBytesReceived: 4,
	}

	query := stats.Query{}

	err := StoreSystemStats(&testRecord)
	if err != nil {
		test.Fatalf("Failed to store stats: '%v'.", err)
	}

	records, err := GetSystemStats(query)
	if err != nil {
		test.Fatalf("Failed to fetch stats: '%v'.", err)
	}

	if len(records) != 1 {
		test.Fatalf("Did not get the correct number of records. Expected: 1, Actual: %d.", len(records))
	}

	if !reflect.DeepEqual(*records[0], testRecord) {
		test.Fatalf("Did not get the expected record back. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(testRecord), util.MustToJSONIndent(*records[0]))
	}
}

func (this *DBTests) DBTestStoreAPIRequestStats(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.APIRequestStatsType,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.SenderKey:       "1",
			stats.EndpointKey:     "E1",
			stats.UserEmailKey:    "U",
			stats.AssignmentIDKey: "A",
			stats.CourseIDKey:     "C",
			stats.LocatorKey:      "11",
		},
	}

	query := stats.Query{
		Type: stats.APIRequestStatsType,
	}

	runStoreStatsTests(test, &metric, query)
}

func (this *DBTests) DBTestStoreTaskTimeStats(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.TaskTimeStatsType,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.UserEmailKey:    "U",
			stats.AssignmentIDKey: "A",
			stats.CourseIDKey:     "C",
		},
	}

	query := stats.Query{
		Type: stats.TaskTimeStatsType,
	}

	runStoreStatsTests(test, &metric, query)
}

func (this *DBTests) DBTestStoreGradingTimeStats(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.GradingTimeStatsType,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.UserEmailKey:    "U",
			stats.AssignmentIDKey: "A",
			stats.CourseIDKey:     "C",
		},
	}

	query := stats.Query{
		Type: stats.GradingTimeStatsType,
	}

	runStoreStatsTests(test, &metric, query)
}

func (this *DBTests) DBTestStoreCodeAnalysisTimeStats(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.CodeAnalysisTimeStatsType,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.UserEmailKey:    "U",
			stats.AssignmentIDKey: "A",
			stats.CourseIDKey:     "C",
		},
	}

	query := stats.Query{
		Type: stats.CodeAnalysisTimeStatsType,
	}

	runStoreStatsTests(test, &metric, query)
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
		Type:      stats.APIRequestStatsType,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.SenderKey:       "1",
			stats.EndpointKey:     "E1",
			stats.UserEmailKey:    "U",
			stats.AssignmentIDKey: "A",
			stats.CourseIDKey:     "C",
			stats.LocatorKey:      "11",
		},
	}

	query := stats.Query{
		Type: stats.APIRequestStatsType,
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

func runStoreStatsTests(test *testing.T, metric *stats.Metric, query stats.Query) {
	err := StoreMetric(metric)
	if err != nil {
		test.Fatalf("Failed to store stats: '%v'.", err)
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
