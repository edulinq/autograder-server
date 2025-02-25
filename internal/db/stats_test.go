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

	testRecord := stats.CourseMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.Now(),
		},
		CourseAssignmentUserMetric: stats.CourseAssignmentUserMetric{
			CourseID:     "C",
			AssignmentID: "A",
			UserEmail:    "U",
		},
		Type:  stats.CourseMetricTypeGradingTime,
		Value: 100,
	}

	query := stats.CourseMetricQuery{
		CourseAssignmentUserQuery: stats.CourseAssignmentUserQuery{
			CourseID: "C",
		},
	}

	err := StoreCourseMetric(&testRecord)
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

func (this *DBTests) DBTestStoreAPIRequestMetrics(test *testing.T) {
	Clear()
	defer Clear()

	testRecord := stats.APIRequestMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.Now(),
		},
		CourseAssignmentUserMetric: stats.CourseAssignmentUserMetric{
			CourseID:     "C",
			AssignmentID: "A",
			UserEmail:    "U",
		},
		Sender:   "2",
		Endpoint: "E",
		Duration: 100,
		Locator:  "1",
	}

	err := StoreAPIRequestMetric(&testRecord)
	if err != nil {
		test.Fatalf("Failed to store stats: '%v'.", err)
	}

	query := stats.APIRequestMetricQuery{}

	records, err := GetAPIRequestMetrics(query)
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
