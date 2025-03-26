package course

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestQuery(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
		email           string
		expectedLocator string
		query           stats.MetricQuery
		expectedResults []map[string]any
	}{
		// Base.
		{
			"server-admin",
			"",
			stats.MetricQuery{},
			[]map[string]any{
				{"assignment": "A1", "course": "course101", "duration": 100, "timestamp": 100, "type": "grading-time", "user": "U1"},
				{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
				{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
			},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{BaseQuery: stats.BaseQuery{Sort: 1}},
			[]map[string]any{
				{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
				{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
				{"assignment": "A1", "course": "course101", "duration": 100, "timestamp": 100, "type": "grading-time", "user": "U1"},
			},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{BaseQuery: stats.BaseQuery{After: timestamp.FromMSecs(150)}},
			[]map[string]any{
				{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
				{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
			},
		},

		// Error.
		{"course-student", "-020", stats.MetricQuery{}, nil},
		{"server-user", "-040", stats.MetricQuery{}, nil},
	}

	for _, record := range testRecords {
		err := db.StoreCourseMetric(record)
		if err != nil {
			test.Fatalf("Failed to store test record: '%v'.", err)
		}
	}

	for i, testCase := range testCases {
		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.query), &fields)

		response := core.SendTestAPIRequestFull(test, `stats/course/query`, fields, nil, testCase.email)
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

		actualSlice := make([]any, len(responseContent.Results))
		for i, data := range responseContent.Results {
			actualSlice[i] = data
		}

		expectedSlice := make([]any, len(testCase.expectedResults))
		for i, data := range testCase.expectedResults {
			expectedSlice[i] = data
		}

		expected := util.MustToGenericJSONSlice(actualSlice, util.JSONCompareFunc)
		actual := util.MustToGenericJSONSlice(expectedSlice, util.JSONCompareFunc)

		if !reflect.DeepEqual(expected, actual) {
			test.Errorf("Case %d: Response is not as expected. Expected: '%v', Actual: '%v'.", i, util.MustToJSONIndent(testCase.expectedResults), util.MustToJSONIndent(responseContent.Results))
			continue
		}
	}
}

var testRecords []*stats.CourseMetric = []*stats.CourseMetric{
	{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(100),
		},
		Type:         stats.CourseMetricTypeGradingTime,
		CourseID:     db.TEST_COURSE_ID,
		AssignmentID: "A1",
		UserEmail:    "U1",
		Value:        100,
	},
	{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(200),
		},
		Type:         stats.CourseMetricTypeUnknown,
		CourseID:     db.TEST_COURSE_ID,
		AssignmentID: "A2",
		UserEmail:    "U1",
		Value:        200,
	},
	{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(300),
		},
		Type:         stats.CourseMetricTypeGradingTime,
		CourseID:     db.TEST_COURSE_ID,
		AssignmentID: "A3",
		UserEmail:    "U2",
		Value:        300,
	},
}
