package apirequest

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

// Since all aggregation is done along the same code path,
// only one metric needs to test aggregation.
func TestAggregation(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
		email           string
		expectedLocator string
		query           stats.MetricQuery
		expectedResults []map[string]any
	}{
		// No group by, numeric aggregation.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{OverviewField: "timestamp"},
			},
			[]map[string]any{
				{
					"group-by": map[string]any{},
					"overview": "timestamp",
					"stats":    map[string]any{"count": 4, "max": 300, "mean": 225, "median": 250, "min": 100},
				},
			},
		},

		// Single group by, numeric aggregation.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{
					GroupByFields: []string{"course"},
					OverviewField: "timestamp",
				},
			},
			[]map[string]any{
				{
					"group-by": map[string]any{"course": "C1"},
					"overview": "timestamp",
					"stats":    map[string]any{"count": 1, "max": 100, "mean": 100, "median": 100, "min": 100},
				},
				{
					"group-by": map[string]any{"course": "C2"},
					"overview": "timestamp",
					"stats":    map[string]any{"count": 1, "max": 200, "mean": 200, "median": 200, "min": 200},
				},
				{
					"group-by": map[string]any{"course": "C3"},
					"overview": "timestamp",
					"stats":    map[string]any{"count": 2, "max": 300, "mean": 300, "median": 300, "min": 300},
				},
			},
		},

		// Multiple group bys, numeric aggregation.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{
					GroupByFields: []string{"course", "assignment"},
					OverviewField: "timestamp",
				},
			},
			[]map[string]any{
				{
					"group-by": map[string]any{"course": "C2", "assignment": "A1"},
					"overview": "timestamp",
					"stats":    map[string]any{"count": 1, "max": 200, "mean": 200, "median": 200, "min": 200},
				},
				{
					"group-by": map[string]any{"course": "C3", "assignment": "A3"},
					"overview": "timestamp",
					"stats":    map[string]any{"count": 2, "max": 300, "mean": 300, "median": 300, "min": 300},
				},
				{
					"group-by": map[string]any{"course": "C1", "assignment": "A1"},
					"overview": "timestamp",
					"stats":    map[string]any{"count": 1, "max": 100, "mean": 100, "median": 100, "min": 100},
				},
			},
		},

		// No group by, non-numeric aggregation.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{OverviewField: "course"},
			},
			[]map[string]any{
				{
					"group-by": map[string]any{},
					"overview": "course",
					"stats":    map[string]any{"count": 4},
				},
			},
		},

		// Single group by, non-numeric aggregation.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{
					GroupByFields: []string{"course"},
					OverviewField: "assignment",
				},
			},
			[]map[string]any{
				{
					"group-by": map[string]any{"course": "C3"},
					"overview": "assignment",
					"stats":    map[string]any{"count": 2},
				},
				{
					"group-by": map[string]any{"course": "C1"},
					"overview": "assignment",
					"stats":    map[string]any{"count": 1},
				},
				{
					"group-by": map[string]any{"course": "C2"},
					"overview": "assignment",
					"stats":    map[string]any{"count": 1},
				},
			},
		},

		// Multiple group bys, non-numeric aggregation.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{
					GroupByFields: []string{"course", "assignment"},
					OverviewField: "locator",
				},
			},
			[]map[string]any{
				{
					"group-by": map[string]any{
						"assignment": "A3",
						"course":     "C3",
					},
					"overview": "locator",
					"stats":    map[string]any{"count": 2, "max": 33, "mean": 33, "median": 33, "min": 33},
				},
				{
					"group-by": map[string]any{
						"assignment": "A1",
						"course":     "C1",
					},
					"overview": "locator",
					"stats":    map[string]any{"count": 1, "max": 11, "mean": 11, "median": 11, "min": 11},
				},
				{
					"group-by": map[string]any{
						"assignment": "A1",
						"course":     "C2",
					},
					"overview": "locator",
					"stats":    map[string]any{"count": 1, "max": 22, "mean": 22, "median": 22, "min": 22},
				},
			},
		},

		// Perm Error.
		{"server-user", "-041", stats.MetricQuery{}, nil},

		// Invalid overview field.
		{"server-admin",
			"-302",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{
					GroupByFields: []string{"course", "assignment"},
					OverviewField: "zzz",
				},
			},
			nil,
		},

		// Invalid group by field.
		{"server-admin",
			"-302",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{
					GroupByFields: []string{"zzz"},
					OverviewField: "assignment",
				},
			},
			nil,
		},

		// Same field in group by and overview field.
		{
			"server-admin",
			"-302",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{
					GroupByFields: []string{"course", "assignment"},
					OverviewField: "assignment",
				},
			},
			nil,
		},

		// Group by field with no overview field.
		{
			"server-admin",
			"-302",
			stats.MetricQuery{
				AggregationQuery: stats.AggregationQuery{
					GroupByFields: []string{"assignment"},
				},
			},
			nil,
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
