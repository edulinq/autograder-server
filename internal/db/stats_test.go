package db

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestStoreSystemStats(test *testing.T) {
	Clear()
	defer Clear()

	testRecord := stats.SystemMetrics{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.Now(),
		},
		CPUPercent:       1,
		MemPercent:       2,
		NetBytesSent:     3,
		NetBytesReceived: 4,
	}

	query := stats.BaseQuery{}

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

func (this *DBTests) DBTestStoreCourseMetrics(test *testing.T) {
	Clear()
	defer Clear()

	testRecord := stats.BaseMetric{
		Timestamp: timestamp.Now(),
		Type:      stats.GRADING_TIME_STATS_KEY,
		Attributes: map[string]any{
			stats.COURSE_ID_KEY:     "C",
			stats.ASSIGNMENT_ID_KEY: "A",
			stats.USER_EMAIL_KEY:    "U",
			stats.VALUE_KEY:         float64(100),
		},
	}

	query := stats.MetricQuery{
		BaseQuery: stats.BaseQuery{
			Where: map[string]any{
				"course": "C",
			},
		},
	}

	err := StoreMetric(&testRecord)
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

	if !reflect.DeepEqual(*records[0], testRecord) {
		test.Fatalf("Did not get the expected record back. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(testRecord), util.MustToJSONIndent(*records[0]))
	}
}

func (this *DBTests) DBTestStoreAPIRequestMetrics(test *testing.T) {
	Clear()
	defer Clear()

	testRecord := stats.BaseMetric{
		Timestamp: timestamp.Now(),
		Type:      stats.API_REQUEST_STATS_KEY,
		Attributes: map[string]any{
			stats.SENDER_KEY:        "2",
			stats.ENDPOINT_KEY:      "E",
			stats.USER_EMAIL_KEY:    "U",
			stats.COURSE_ID_KEY:     "C",
			stats.ASSIGNMENT_ID_KEY: "A",
			stats.LOCATOR_KEY:       "1",
			stats.DURATION_KEY:      float64(100),
		},
	}

	err := StoreMetric(&testRecord)
	if err != nil {
		test.Fatalf("Failed to store stats: '%v'.", err)
	}

	query := stats.MetricQuery{
		BaseQuery: stats.BaseQuery{
			Type: stats.API_REQUEST_STATS_KEY,
		},
	}

	records, err := GetMetrics(query)
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
