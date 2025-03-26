package apirequest

import (
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
		expectedValues  []int
	}{
		// Base
		{"server-admin", "", stats.MetricQuery{}, []int{100, 200, 300}},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Sort: 1}}, []int{300, 200, 100}},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{After: timestamp.FromMSecs(150)}}, []int{200, 300}},

		// Course Specifc
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.ASSIGNMENT_ID: "A2"}}}, []int{200}},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.ASSIGNMENT_ID: "zzz"}}}, nil},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.USER_EMAIL: "U1"}}}, []int{100, 200}},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.USER_EMAIL: "zzz"}}}, nil},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.COURSE_ID: "C1"}}}, []int{100}},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.COURSE_ID: "zzz"}}}, nil},

		// Endpoint Specific
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.ENDPOINT: "E1"}}}, []int{100}},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.ENDPOINT: "zzz"}}}, nil},

		// Sender Specific
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.SENDER: "1"}}}, []int{100}},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.SENDER: "zzz"}}}, nil},

		// Locator Specific
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.LOCATOR: "11"}}}, []int{100}},
		{"server-admin", "", stats.MetricQuery{BaseQuery: stats.BaseQuery{Where: map[string]any{stats.LOCATOR: "zzz"}}}, nil},
		// Error
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

var testRecords []*stats.BaseMetric = []*stats.BaseMetric{
	&stats.BaseMetric{
		Timestamp: timestamp.FromMSecs(100),
		Attributes: map[string]any{
			stats.SENDER:        "1",
			stats.ENDPOINT:      "E1",
			stats.USER_EMAIL:    "U1",
			stats.ASSIGNMENT_ID: "A1",
			stats.COURSE_ID:     "C1",
			stats.LOCATOR:       "11",
			stats.DURATION:      100,
		},
	},
	&stats.BaseMetric{
		Timestamp: timestamp.FromMSecs(200),
		Attributes: map[string]any{
			stats.SENDER:        "2",
			stats.ENDPOINT:      "E2",
			stats.USER_EMAIL:    "U1",
			stats.COURSE_ID:     "C2",
			stats.ASSIGNMENT_ID: "A2",
			stats.LOCATOR:       "22",
			stats.DURATION:      200,
		},
	},
	&stats.BaseMetric{
		Timestamp: timestamp.FromMSecs(300),
		Attributes: map[string]any{
			stats.SENDER:        "3",
			stats.ENDPOINT:      "E3",
			stats.USER_EMAIL:    "U3",
			stats.COURSE_ID:     "C3",
			stats.ASSIGNMENT_ID: "A3",
			stats.LOCATOR:       "33",
			stats.DURATION:      300,
		},
	},
}
