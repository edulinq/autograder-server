package apirequest

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// Since all aggregation is done along the same code path,
// only one metric needs to test aggregation.
func TestQuery(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
		email           string
		expectedLocator string
		query           stats.APIRequestMetricQuery
		expectedResults []map[string]any
	}{
		// Base.
		{"server-admin", "", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{}}, []map[string]any{
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{BaseQuery: stats.BaseQuery{Sort: 1}}, []map[string]any{
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{BaseQuery: stats.BaseQuery{After: timestamp.FromMSecs(150)}}, []map[string]any{
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
		}},

		// Include.
		{"server-admin", "", stats.APIRequestMetricQuery{IncludeAPIRequestMetricField: stats.IncludeAPIRequestMetricField{Sender: "1"}}, []map[string]any{
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{IncludeAPIRequestMetricField: stats.IncludeAPIRequestMetricField{Endpoint: "E1"}}, []map[string]any{
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{IncludeAPIRequestMetricField: stats.IncludeAPIRequestMetricField{UserEmail: "U1"}}, []map[string]any{
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{IncludeAPIRequestMetricField: stats.IncludeAPIRequestMetricField{CourseID: "C1"}}, []map[string]any{
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{IncludeAPIRequestMetricField: stats.IncludeAPIRequestMetricField{AssignmentID: "A1"}}, []map[string]any{
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{IncludeAPIRequestMetricField: stats.IncludeAPIRequestMetricField{Locator: "11"}}, []map[string]any{
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{IncludeAPIRequestMetricField: stats.IncludeAPIRequestMetricField{Locator: "ZZZ"}}, []map[string]any{}},

		// Exclude.
		{"server-admin", "", stats.APIRequestMetricQuery{ExcludeAPIRequestMetricField: stats.ExcludeAPIRequestMetricField{Sender: "1"}}, []map[string]any{
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{ExcludeAPIRequestMetricField: stats.ExcludeAPIRequestMetricField{Endpoint: "E1"}}, []map[string]any{
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{ExcludeAPIRequestMetricField: stats.ExcludeAPIRequestMetricField{UserEmail: "U1"}}, []map[string]any{
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{ExcludeAPIRequestMetricField: stats.ExcludeAPIRequestMetricField{CourseID: "C1"}}, []map[string]any{
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{ExcludeAPIRequestMetricField: stats.ExcludeAPIRequestMetricField{AssignmentID: "A1"}}, []map[string]any{
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{ExcludeAPIRequestMetricField: stats.ExcludeAPIRequestMetricField{Locator: "11"}}, []map[string]any{
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
		}},
		{"server-admin", "", stats.APIRequestMetricQuery{ExcludeAPIRequestMetricField: stats.ExcludeAPIRequestMetricField{Locator: "ZZZ"}}, []map[string]any{
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
			{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
		}},

		// Include and Exclude different fields.
		{"server-admin", "", stats.APIRequestMetricQuery{IncludeAPIRequestMetricField: stats.IncludeAPIRequestMetricField{AssignmentID: "A1"}, ExcludeAPIRequestMetricField: stats.ExcludeAPIRequestMetricField{Endpoint: "E2"}}, []map[string]any{
			{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
		}},
		// Include and Exclude same fields.
		{"server-admin", "", stats.APIRequestMetricQuery{IncludeAPIRequestMetricField: stats.IncludeAPIRequestMetricField{CourseID: "C1"}, ExcludeAPIRequestMetricField: stats.ExcludeAPIRequestMetricField{CourseID: "C1"}}, nil},

		// Aggregation.
		// No group by, numeric aggregation.
		{"server-admin", "", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, AggregateField: "timestamp"}}, []map[string]any{
			{"count": 4, "max": 300, "mean": 225, "median": 250, "min": 100},
		}},

		// Single group by, numeric aggregation.
		{"server-admin", "", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, GroupByFields: []string{"course"}, AggregateField: "timestamp"}}, []map[string]any{
			{"count": 1, "course": "C1", "max": 100, "mean": 100, "median": 100, "min": 100},
			{"count": 1, "course": "C2", "max": 200, "mean": 200, "median": 200, "min": 200},
			{"count": 2, "course": "C3", "max": 300, "mean": 300, "median": 300, "min": 300},
		}},

		// Multiple group bys, numeric aggregation.
		{"server-admin", "", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, GroupByFields: []string{"course", "assignment"}, AggregateField: "timestamp"}}, []map[string]any{
			{"assignment": "A1", "count": 1, "course": "C2", "max": 200, "mean": 200, "median": 200, "min": 200},
			{"assignment": "A3", "count": 2, "course": "C3", "max": 300, "mean": 300, "median": 300, "min": 300},
			{"assignment": "A1", "count": 1, "course": "C1", "max": 100, "mean": 100, "median": 100, "min": 100},
		}},

		// No group by, non-numeric aggregation.
		{"server-admin", "", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, AggregateField: "course"}}, []map[string]any{
			{"count": 4},
		}},

		// Single group by, non-numeric aggregation.
		{"server-admin", "", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, GroupByFields: []string{"course"}, AggregateField: "assignment"}}, []map[string]any{
			{"count": 2, "course": "C3"},
			{"count": 1, "course": "C1"},
			{"count": 1, "course": "C2"},
		}},

		// Multiple group bys, non-numeric aggregation.
		{"server-admin", "", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, GroupByFields: []string{"course", "assignment"}, AggregateField: "assignment"}}, []map[string]any{
			{"assignment": "A3", "count": 2, "course": "C3"},
			{"assignment": "A1", "count": 1, "course": "C1"},
			{"assignment": "A1", "count": 1, "course": "C2"},
		}},

		// Error.
		{"server-user", "-041", stats.APIRequestMetricQuery{}, nil},
		{"server-admin", "-303", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, GroupByFields: []string{"course", "assignment"}}}, nil},
		{"server-admin", "-304", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, GroupByFields: []string{"course", "assignment"}, AggregateField: "zzz"}}, nil},
		{"server-admin", "-304", stats.APIRequestMetricQuery{AggregationQuery: stats.AggregationQuery{EnableAggregation: true, GroupByFields: []string{"zzz"}, AggregateField: "assignment"}}, nil},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		for _, record := range testRecords {
			err := db.StoreAPIRequestMetric(record)
			if err != nil {
				test.Fatalf("Failed to store test record: '%v'.", err)
			}
		}

		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.query), &fields)

		response := core.SendTestAPIRequestFull(test, `stats/apirequest/query`, fields, nil, testCase.email)
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

var testRecords []*stats.APIRequestMetric = []*stats.APIRequestMetric{
	&stats.APIRequestMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(100),
		},
		Sender:       "1",
		Endpoint:     "E1",
		UserEmail:    "U1",
		AssignmentID: "A1",
		CourseID:     "C1",
		Locator:      "11",
		Duration:     100,
	},
	&stats.APIRequestMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(200),
		},
		Sender:       "2",
		Endpoint:     "E2",
		UserEmail:    "U2",
		CourseID:     "C2",
		AssignmentID: "A1",
		Locator:      "22",
		Duration:     200,
	},
	&stats.APIRequestMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(300),
		},
		Sender:       "3",
		Endpoint:     "E3",
		UserEmail:    "U3",
		CourseID:     "C3",
		AssignmentID: "A3",
		Locator:      "33",
		Duration:     300,
	},
	&stats.APIRequestMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(300),
		},
		Sender:       "3",
		Endpoint:     "E3",
		UserEmail:    "U3",
		CourseID:     "C3",
		AssignmentID: "A3",
		Locator:      "33",
		Duration:     300,
	},
}
