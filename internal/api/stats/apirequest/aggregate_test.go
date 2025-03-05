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

func TestAggregate(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		email           string
		expectedLocator string
		query           stats.APIRequestMetricAggregate
		expected        *map[string]util.AggregateValues
	}{
		// Single Include.
		{
			email: "server-admin",
			query: stats.APIRequestMetricAggregate{
				GroupBy: stats.UserEmail,
				Filters: map[stats.APIRequestFieldType]stats.Filter{
					stats.UserEmail: {
						Include: []string{"U1"},
					},
				},
			},
			expected: &map[string]util.AggregateValues{
				"U1": {Count: 2, Mean: 150, Median: 150, Min: 100, Max: 200},
			},
		},

		// Single Include Does Not Exist.
		{
			email: "server-admin",
			query: stats.APIRequestMetricAggregate{
				GroupBy: stats.UserEmail,
				Filters: map[stats.APIRequestFieldType]stats.Filter{
					stats.UserEmail: {
						Include: []string{"ZZZ"},
					},
				},
			},
			expected: &map[string]util.AggregateValues{},
		},

		// Single Include with Base Query.
		{
			email: "server-admin",
			query: stats.APIRequestMetricAggregate{
				BaseQuery: stats.BaseQuery{After: timestamp.FromMSecs(100)},
				GroupBy:   stats.UserEmail,
				Filters: map[stats.APIRequestFieldType]stats.Filter{
					stats.UserEmail: {
						Include: []string{"U1"},
					},
				},
			},
			expected: &map[string]util.AggregateValues{
				"U1": {Count: 1, Mean: 200, Median: 200, Min: 200, Max: 200},
			},
		},

		// Multiple Includes.
		{
			email: "server-admin",
			query: stats.APIRequestMetricAggregate{
				GroupBy: stats.UserEmail,
				Filters: map[stats.APIRequestFieldType]stats.Filter{
					stats.UserEmail: {
						Include: []string{"U1", "U3"},
					},
				},
			},
			expected: &map[string]util.AggregateValues{
				"U1": {Count: 2, Mean: 150, Median: 150, Min: 100, Max: 200},
				"U3": {Count: 1, Mean: 300, Median: 300, Min: 300, Max: 300},
			},
		},

		// Single Exclude.
		{
			email: "server-admin",
			query: stats.APIRequestMetricAggregate{
				GroupBy: stats.UserEmail,
				Filters: map[stats.APIRequestFieldType]stats.Filter{
					stats.UserEmail: {
						Exclude: []string{"U1"},
					},
				},
			},
			expected: &map[string]util.AggregateValues{
				"U3": {Count: 1, Mean: 300, Median: 300, Min: 300, Max: 300},
			},
		},

		// Multiple Excludes.
		{
			email: "server-admin",
			query: stats.APIRequestMetricAggregate{
				GroupBy: stats.UserEmail,
				Filters: map[stats.APIRequestFieldType]stats.Filter{
					stats.UserEmail: {
						Exclude: []string{"U1", "U3"},
					},
				},
			},
			expected: &map[string]util.AggregateValues{},
		},

		// Include and Exclude.
		{
			email: "server-admin",
			query: stats.APIRequestMetricAggregate{
				GroupBy: stats.UserEmail,
				Filters: map[stats.APIRequestFieldType]stats.Filter{
					stats.UserEmail: {
						Include: []string{"U1"},
					},
					stats.CourseID: {
						Exclude: []string{"C2"},
					},
				},
			},
			expected: &map[string]util.AggregateValues{
				"U1": {Count: 1, Mean: 100, Median: 100, Min: 100, Max: 100},
			},
		},

		// No Filters.
		{
			email: "server-admin",
			query: stats.APIRequestMetricAggregate{
				GroupBy: stats.UserEmail,
			},
			expected: &map[string]util.AggregateValues{
				"U1": {Count: 2, Mean: 150, Median: 150, Min: 100, Max: 200},
				"U3": {Count: 1, Mean: 300, Median: 300, Min: 300, Max: 300},
			},
		},

		// Error.
		{
			email:           "server-user",
			expectedLocator: "-041",
			query:           stats.APIRequestMetricAggregate{},
			expected:        nil,
		},
		{
			email:           "server-admin",
			expectedLocator: "-302",
			query:           stats.APIRequestMetricAggregate{},
			expected:        nil,
		},
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

		response := core.SendTestAPIRequestFull(test, `stats/apirequest/aggregate`, fields, nil, testCase.email)
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

		var responseContent AggregateResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(testCase.expected, responseContent.Records) {
			test.Errorf("Case %d: Response is not as expected. Expected: '%v', Actual: '%v'.", i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent.Records))
			continue
		}
	}
}
