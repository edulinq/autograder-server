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
		email          string
		query          *stats.Query
		expectedValues []int
		permError      bool
	}{
		// Grading Time Stats.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: db.TEST_COURSE_ID,
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "zzz",
				},
			},
			expectedValues: []int{100},
		},

		// Task Time Stats.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: db.TEST_COURSE_ID,
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "zzz",
				},
			},
			expectedValues: []int{100},
		},

		// Code Analysis Time Stats
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: db.TEST_COURSE_ID,
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "zzz",
				},
			},
			expectedValues: []int{100},
		},

		// Perm Error
		{
			email: "course-student",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
			},
			permError: true,
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
			if testCase.permError {
				expectedLocator := "-020"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.permError {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
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
		Type:      stats.MetricTypeGradingTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     db.TEST_COURSE_ID,
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     db.TEST_COURSE_ID,
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.MetricTypeCodeAnalysisTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     db.TEST_COURSE_ID,
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},

	// Non-context course metrics.
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.MetricTypeGradingTime,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C1",
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C1",
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.MetricTypeCodeAnalysisTime,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C1",
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
}
