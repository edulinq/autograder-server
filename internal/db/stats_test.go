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

func (this *DBTests) DBTestStoreMetrics(test *testing.T) {
	Clear()
	defer Clear()

	testCases := []struct {
		Metric                stats.Metric
		Query                 stats.Query
		ExpectedRecordsNumber int
		ExpectedError         bool
	}{
		// API Request Stats.
		{
			Metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Type:      stats.API_REQUEST_STATS_TYPE,
				Attributes: map[stats.MetricAttribute]any{
					stats.SENDER_KEY:        "1",
					stats.ENDPOINT_KEY:      "E1",
					stats.USER_EMAIL_KEY:    "U",
					stats.ASSIGNMENT_ID_KEY: "A",
					stats.COURSE_ID_KEY:     "C",
					stats.LOCATOR_KEY:       "11",
					stats.DURATION_KEY:      float64(100),
				},
			},
			Query: stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
			ExpectedRecordsNumber: 1,
		},

		// Task Time Stats.
		{
			Metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Type:      stats.TASK_TIME_STATS_TYPE,
				Attributes: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY:     "C",
					stats.ASSIGNMENT_ID_KEY: "A",
					stats.USER_EMAIL_KEY:    "U",
					stats.DURATION_KEY:      float64(100),
				},
			},
			Query: stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
			ExpectedRecordsNumber: 1,
		},

		// Grading Time Stats.
		{
			Metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Type:      stats.GRADING_TIME_STATS_TYPE,
				Attributes: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY:     "C",
					stats.ASSIGNMENT_ID_KEY: "A",
					stats.USER_EMAIL_KEY:    "U",
					stats.DURATION_KEY:      float64(100),
				},
			},
			Query: stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
			ExpectedRecordsNumber: 1,
		},

		// Code Analysis Time Stats.
		{
			Metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Attributes: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY:     "C",
					stats.ASSIGNMENT_ID_KEY: "A",
					stats.USER_EMAIL_KEY:    "U",
					stats.DURATION_KEY:      float64(100),
				},
			},
			Query: stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
			ExpectedRecordsNumber: 1,
		},

		// Mixed Metric and Query Type.
		{
			Metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Attributes: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY:     "C",
					stats.ASSIGNMENT_ID_KEY: "A",
					stats.USER_EMAIL_KEY:    "U",
					stats.DURATION_KEY:      float64(100),
				},
			},
			Query: stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
		},

		// No Metric Type.
		{
			Metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Attributes: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY:     "C",
					stats.ASSIGNMENT_ID_KEY: "A",
					stats.USER_EMAIL_KEY:    "U",
					stats.DURATION_KEY:      float64(100),
				},
			},
			Query: stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
			ExpectedError: true,
		},

		// No Query Type.
		{
			Metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Attributes: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY:     "C",
					stats.ASSIGNMENT_ID_KEY: "A",
					stats.USER_EMAIL_KEY:    "U",
					stats.DURATION_KEY:      float64(100),
				},
			},
			Query: stats.Query{
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
			ExpectedError: true,
		},
	}

	for i, testCase := range testCases {
		Clear()

		err := StoreMetric(&testCase.Metric)
		if err != nil && !testCase.ExpectedError {
			test.Errorf("Case %d: Failed to store stats: '%v'.", i, err)
			continue
		}

		if err == nil && testCase.ExpectedError {
			test.Errorf("Case %d: Unexpected success when storing metric '%v'.", i, util.MustToJSONIndent(testCase.Metric))
		}

		if testCase.ExpectedError {
			continue
		}

		records, err := GetMetrics(testCase.Query)
		if err != nil {
			test.Errorf("Case %d: Failed to fetch stats: '%v'.", i, err)
			continue
		}

		if len(records) != testCase.ExpectedRecordsNumber {
			test.Errorf("Case %d: Did not get the correct number of records. Expected: %d, Actual: %d.", i, testCase.ExpectedRecordsNumber, len(records))
			continue
		}

		if testCase.ExpectedRecordsNumber == 0 {
			continue
		}

		if !reflect.DeepEqual(*records[0], testCase.Metric) {
			test.Errorf("Case %d: Did not get the expected record back. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.Metric), util.MustToJSONIndent(*records[0]))
			continue
		}
	}
}
