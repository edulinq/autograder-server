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
		query           stats.CourseMetricQuery
		expectedResults []map[string]any
	}{
		// Base.
		{"server-admin", "", stats.CourseMetricQuery{}, []map[string]any{
			{"assignment": "A1", "course": "course101", "duration": 100, "timestamp": 100, "type": "grading-time", "user": "U1"},
			{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
			{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
		}},
		{"server-admin", "", stats.CourseMetricQuery{BaseQuery: stats.BaseQuery{Sort: 1}}, []map[string]any{
			{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
			{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
			{"assignment": "A1", "course": "course101", "duration": 100, "timestamp": 100, "type": "grading-time", "user": "U1"},
		}},
		{"server-admin", "", stats.CourseMetricQuery{BaseQuery: stats.BaseQuery{After: timestamp.FromMSecs(150)}}, []map[string]any{
			{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
			{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
		}},

		// Include.
		{"server-admin", "", stats.CourseMetricQuery{IncludeCourseMetricField: stats.IncludeCourseMetricField{AssignmentID: "A2"}}, []map[string]any{
			{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
		}},
		{"server-admin", "", stats.CourseMetricQuery{IncludeCourseMetricField: stats.IncludeCourseMetricField{AssignmentID: "ZZZ"}}, nil},
		{"server-admin", "", stats.CourseMetricQuery{IncludeCourseMetricField: stats.IncludeCourseMetricField{UserEmail: "U2"}}, []map[string]any{
			{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
		}},
		{"server-admin", "", stats.CourseMetricQuery{IncludeCourseMetricField: stats.IncludeCourseMetricField{Type: stats.CourseMetricTypeGradingTime}}, []map[string]any{
			{"assignment": "A1", "course": "course101", "duration": 100, "timestamp": 100, "type": "grading-time", "user": "U1"},
			{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
		}},

		// Exclude.
		{"server-admin", "", stats.CourseMetricQuery{ExcludeCourseMetricField: stats.ExcludeCourseMetricField{AssignmentID: "A2"}}, []map[string]any{
			{"assignment": "A1", "course": "course101", "duration": 100, "timestamp": 100, "type": "grading-time", "user": "U1"},
			{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
		}},
		{"server-admin", "", stats.CourseMetricQuery{ExcludeCourseMetricField: stats.ExcludeCourseMetricField{AssignmentID: "ZZZ"}}, []map[string]any{
			{"assignment": "A1", "course": "course101", "duration": 100, "timestamp": 100, "type": "grading-time", "user": "U1"},
			{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
			{"assignment": "A3", "course": "course101", "duration": 300, "timestamp": 300, "type": "grading-time", "user": "U2"},
		}},
		{"server-admin", "", stats.CourseMetricQuery{ExcludeCourseMetricField: stats.ExcludeCourseMetricField{UserEmail: "U2"}}, []map[string]any{
			{"assignment": "A1", "course": "course101", "duration": 100, "timestamp": 100, "type": "grading-time", "user": "U1"},
			{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
		}},
		{"server-admin", "", stats.CourseMetricQuery{ExcludeCourseMetricField: stats.ExcludeCourseMetricField{Type: stats.CourseMetricTypeGradingTime}}, []map[string]any{
			{"assignment": "A2", "course": "course101", "duration": 200, "timestamp": 200, "type": "", "user": "U1"},
		}},

		// Include and Exclude different fields.
		{"server-admin", "", stats.CourseMetricQuery{IncludeCourseMetricField: stats.IncludeCourseMetricField{UserEmail: "U1"}, ExcludeCourseMetricField: stats.ExcludeCourseMetricField{AssignmentID: "A2"}}, []map[string]any{
			{"assignment": "A1", "course": "course101", "duration": 100, "timestamp": 100, "type": "grading-time", "user": "U1"},
		}},

		// Include and Exclude same fields.
		{"server-admin", "", stats.CourseMetricQuery{IncludeCourseMetricField: stats.IncludeCourseMetricField{UserEmail: "U1"}, ExcludeCourseMetricField: stats.ExcludeCourseMetricField{UserEmail: "U1"}}, []map[string]any{}},

		// Error.
		{"course-student", "-020", stats.CourseMetricQuery{}, nil},
		{"server-user", "-040", stats.CourseMetricQuery{}, nil},
		{"server-admin", "-307", stats.CourseMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, GroupByFields: []string{"course", "assignment"}}}, nil},
		{"server-admin", "-308", stats.CourseMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, GroupByFields: []string{"zzz"}, AggregateField: "assignment"}}, nil},
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

		var responseContent stats.QueryResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		actualSlice := make([]any, len(responseContent.Response))
		for i, data := range responseContent.Response {
			actualSlice[i] = data
		}

		expectedSlice := make([]any, len(testCase.expectedResults))
		for i, data := range testCase.expectedResults {
			expectedSlice[i] = data
		}

		expected := util.MustToGenericJSON(actualSlice, stats.QuerySortFuncForTesting)
		actual := util.MustToGenericJSON(expectedSlice, stats.QuerySortFuncForTesting)

		if !reflect.DeepEqual(expected, actual) {
			test.Errorf("Case %d: Response is not as expected. Expected: '%v', Actual: '%v'.", i, util.MustToJSONIndent(testCase.expectedResults), util.MustToJSONIndent(responseContent.Response))
			continue
		}
	}
}

var testRecords []*stats.CourseMetric = []*stats.CourseMetric{
	&stats.CourseMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(100),
		},
		Type:         stats.CourseMetricTypeGradingTime,
		CourseID:     db.TEST_COURSE_ID,
		AssignmentID: "A1",
		UserEmail:    "U1",
		Value:        100,
	},
	&stats.CourseMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(200),
		},
		Type:         stats.CourseMetricTypeUnknown,
		CourseID:     db.TEST_COURSE_ID,
		AssignmentID: "A2",
		UserEmail:    "U1",
		Value:        200,
	},
	&stats.CourseMetric{
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
