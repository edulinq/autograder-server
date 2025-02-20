package request

import (
	"fmt"
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
		email          string
		permError      bool
		query          stats.BaseQuery
		expectedValues []int
	}{
		{"server-admin", false, stats.BaseQuery{}, []int{100, 200, 300}},
		{"server-admin", false, stats.BaseQuery{Sort: 1}, []int{300, 200, 100}},
		{"server-admin", false, stats.BaseQuery{After: timestamp.FromMSecs(150)}, []int{200, 300}},

		{"server-user", true, stats.BaseQuery{}, []int{}},
	}

	for _, record := range testRecords {
		err := db.StoreRequestMetric(record)
		if err != nil {
			test.Fatalf("Failed to store test record: '%v'.", err)
		}
	}

	for i, testCase := range testCases {
		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.query), &fields)

		response := core.SendTestAPIRequestFull(test, `stats/request/query`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-041"
				if expectedLocator != response.Locator {
					test.Errorf("Case %d: Incorrect locator on perm error. Expected: '%s', Actual: '%s'.", i, expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		var responseContent QueryResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if len(testCase.expectedValues) != len(responseContent.Records) {
			fmt.Println("response: ", util.MustToJSONIndent(responseContent))
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

var testRecords []*stats.RequestMetric = []*stats.RequestMetric{
	&stats.RequestMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(100),
		},
		Sender:       "1",
		Endpoint:     "E1",
		Duration:     100,
		CourseID:     "C1",
		AssignmentID: "A1",
		UserEmail:    "U1",
		Locator:      "11",
	},
	&stats.RequestMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(200),
		},
		Sender:       "2",
		Endpoint:     "E2",
		Duration:     200,
		CourseID:     "C2",
		AssignmentID: "A2",
		UserEmail:    "U2",
		Locator:      "22",
	},
	&stats.RequestMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(300),
		},
		Sender:       "3",
		Endpoint:     "E3",
		Duration:     300,
		CourseID:     "C3",
		AssignmentID: "A3",
		UserEmail:    "U3",
		Locator:      "33",
	},
}
