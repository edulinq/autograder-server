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
			stats.MetricQuery{AggregationQuery: stats.AggregationQuery{}},
			[]map[string]any{
				{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
				{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
				{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
				{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{BaseQuery: stats.BaseQuery{Sort: 1}},
			[]map[string]any{
				{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
				{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
				{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
				{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
			},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{BaseQuery: stats.BaseQuery{After: timestamp.FromMSecs(150)}},
			[]map[string]any{
				{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
				{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
				{"assignment": "A3", "course": "C3", "duration": 300, "endpoint": "E3", "locator": "33", "sender": "3", "timestamp": 300, "user": "U3"},
			},
		},

		// Include one field.
		{
			"server-admin",
			"",
			stats.MetricQuery{Where: map[string]string{"sender": "1"}},
			[]map[string]any{
				{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
			},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{Where: map[string]string{"endpoint": "E1"}},
			[]map[string]any{
				{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
			},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{Where: map[string]string{"user": "U1"}},
			[]map[string]any{
				{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
			},
		},
		{"server-admin",
			"",
			stats.MetricQuery{Where: map[string]string{"course": "C1"}},
			[]map[string]any{
				{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
			},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{Where: map[string]string{"assignment": "A1"}},
			[]map[string]any{
				{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
				{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{Where: map[string]string{"locator": "11"}},
			[]map[string]any{
				{"assignment": "A1", "course": "C1", "duration": 100, "endpoint": "E1", "locator": "11", "sender": "1", "timestamp": 100, "user": "U1"},
			},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{Where: map[string]string{"locator": "ZZZ"}},
			nil,
		},

		// Include multiple fields.
		{
			"server-admin",
			"",
			stats.MetricQuery{Where: map[string]string{"assignment": "A1", "course": "C2"}},
			[]map[string]any{
				{"assignment": "A1", "course": "C2", "duration": 200, "endpoint": "E2", "locator": "22", "sender": "2", "timestamp": 200, "user": "U2"},
			},
		},

		// No aggregation field.
		{
			"server-admin",
			"-302",
			stats.MetricQuery{AggregationQuery: stats.AggregationQuery{GroupByFields: []string{"course"}}},
			nil,
		},

		// Error.
		{"course-student", "-041", stats.MetricQuery{}, nil},
		{"server-user", "-041", stats.MetricQuery{}, nil},
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

var testRecords []*stats.APIRequestMetric = []*stats.APIRequestMetric{
	{
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
	{
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
	{
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
	{
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
