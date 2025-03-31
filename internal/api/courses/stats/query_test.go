package stats

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// Test that the context course is correctly set when quering stats.
func TestQueryContextCourse(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
		email           string
		expectedLocator string
		query           *stats.Query
		expectedValues  []int
	}{
		// Grading Time Stats.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: db.TEST_COURSE_ID,
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "zzz",
				},
			},
			expectedValues: []int{100},
		},

		// Task Time Stats.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: db.TEST_COURSE_ID,
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "zzz",
				},
			},
			expectedValues: []int{100},
		},

		// Code Analysis Time Stats
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: db.TEST_COURSE_ID,
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "zzz",
				},
			},
			expectedValues: []int{100},
		},

		// Grading Time Stats, Error
		{
			email:           "course-user",
			expectedLocator: "-017",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
			},
		},

		// Task Time Stats, Error
		{
			email:           "course-user",
			expectedLocator: "-017",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
			},
		},

		// Code Analysis Time Stats, Error
		{
			email:           "course-user",
			expectedLocator: "-017",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
			},
		},
	}

	for _, record := range testRecords {
		err := db.StoreMetric(record)
		if err != nil {
			test.Fatalf("Failed to store test record: '%v'.", err)
		}
	}

	for i, testCase := range testCases {
		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.query), &fields)

		response := core.SendTestAPIRequestFull(test, `courses/stats/query`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.expectedLocator != "" {
				if testCase.expectedLocator != response.Locator {
					test.Errorf("Case %d: Incorrect locator. Expected: '%s', Actual: '%s'.", i, testCase.expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.expectedLocator != "" {
			test.Errorf("Case %d: Unexpected success when locator '%s' was expected.", i, testCase.expectedLocator)
			continue
		}

		var responseContent QueryResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if len(testCase.expectedValues) != len(responseContent.Records) {
			test.Errorf("Case %d: Unexpected number of records. Expected: %d, Actual: %d.", i, len(testCase.expectedValues), len(responseContent.Records))
			continue
		}

		match := true
		for i, _ := range responseContent.Records {
			expectedTimestamp := timestamp.FromMSecs(int64(testCase.expectedValues[i]))
			match = (match && (expectedTimestamp == responseContent.Records[i].Timestamp))
		}

		if !match {
			test.Errorf("Case %d: Unexpected record timestamps. Expected: %s, Actual: %s.", i, util.MustToJSONIndent(testCase.expectedValues), util.MustToJSONIndent(responseContent.Records))
			continue
		}
	}
}

var testRecords []*stats.Metric = []*stats.Metric{
	// Context course metrics.
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.Grading_Time_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     db.TEST_COURSE_ID,
			stats.Assignment_ID_Key: "A1",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.Task_Time_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     db.TEST_COURSE_ID,
			stats.Assignment_ID_Key: "A1",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.Code_Analysis_Time_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     db.TEST_COURSE_ID,
			stats.Assignment_ID_Key: "A1",
			stats.User_Email_Key:    "U1",
		},
	},

	// Non-context course metrics.
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.Grading_Time_Stats_Type,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C1",
			stats.Assignment_ID_Key: "A1",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.Task_Time_Stats_Type,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C1",
			stats.Assignment_ID_Key: "A1",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.Code_Analysis_Time_Stats_Type,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C1",
			stats.Assignment_ID_Key: "A1",
			stats.User_Email_Key:    "U1",
		},
	},
}
