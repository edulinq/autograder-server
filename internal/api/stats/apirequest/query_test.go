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
		query           stats.APIRequestMetricQuery
		expectedValues  []int
	}{
		// Base
		{"server-admin", "", stats.APIRequestMetricQuery{}, []int{100, 200, 300}},
		{"server-admin", "", stats.APIRequestMetricQuery{BaseQuery: stats.BaseQuery{Sort: 1}}, []int{300, 200, 100}},
		{"server-admin", "", stats.APIRequestMetricQuery{BaseQuery: stats.BaseQuery{After: timestamp.FromMSecs(150)}}, []int{200, 300}},

		// Course Specifc
		{"server-admin", "", stats.APIRequestMetricQuery{AssignmentID: "A2"}, []int{200}},
		{"server-admin", "", stats.APIRequestMetricQuery{AssignmentID: "zzz"}, nil},
		{"server-admin", "", stats.APIRequestMetricQuery{UserEmail: "U1"}, []int{100, 200}},
		{"server-admin", "", stats.APIRequestMetricQuery{UserEmail: "zzz"}, nil},
		{"server-admin", "", stats.APIRequestMetricQuery{CourseID: "C1"}, []int{100}},
		{"server-admin", "", stats.APIRequestMetricQuery{CourseID: "zzz"}, nil},

		// Endpoint Specific
		{"server-admin", "", stats.APIRequestMetricQuery{Endpoint: "E1"}, []int{100}},
		{"server-admin", "", stats.APIRequestMetricQuery{Endpoint: "zzz"}, nil},

		// Sender Specific
		{"server-admin", "", stats.APIRequestMetricQuery{Sender: "1"}, []int{100}},
		{"server-admin", "", stats.APIRequestMetricQuery{Sender: "zzz"}, nil},

		// Locator Specific
		{"server-admin", "", stats.APIRequestMetricQuery{Locator: "11"}, []int{100}},
		{"server-admin", "", stats.APIRequestMetricQuery{Locator: "zzz"}, nil},

		// Error
		{"server-user", "-041", stats.APIRequestMetricQuery{}, nil},
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
		UserEmail:    "U1",
		CourseID:     "C2",
		AssignmentID: "A2",
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
}
