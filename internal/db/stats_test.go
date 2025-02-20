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

	err := backend.StoreSystemStats(&testRecord)
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

	testRecord := stats.CourseMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.Now(),
		},
		Type:         stats.CourseMetricTypeGradingTime,
		CourseID:     "C",
		AssignmentID: "A",
		UserEmail:    "U",
		Value:        100,
	}

	query := stats.CourseMetricQuery{
		CourseID: "C",
	}

	err := backend.StoreCourseMetric(&testRecord)
	if err != nil {
		test.Fatalf("Failed to store stats: '%v'.", err)
	}

	records, err := GetCourseMetrics(query)
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

func (this *DBTests) DBTestStoreRequestMetrics(test *testing.T) {
	Clear()
	defer Clear()

	testRecord := stats.RequestMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.Now(),
		},
		CourseID:     "C",
		AssignmentID: "A",
		UserEmail:    "U",
		Endpoint:     "E",
		Locator:      "1",
		IPAddress:    "2",
		Value:        100,
	}

	err := backend.StoreRequestMetric(&testRecord)
	if err != nil {
		test.Fatalf("Failed to store stats: '%v'.", err)
	}
}
