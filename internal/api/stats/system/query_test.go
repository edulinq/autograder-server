package system

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
		email          string
		permError      bool
		query          stats.BaseQuery
		expectedValues []int
	}{
		{"server-admin", false, stats.BaseQuery{}, []int{100, 200, 300}},
		{"server-admin", false, stats.BaseQuery{Sort: 1}, []int{300, 200, 100}},
		{"server-admin", false, stats.BaseQuery{After: timestamp.FromMSecs(150)}, []int{200, 300}},

		{"server-user", true, stats.BaseQuery{}, nil},
	}

	for _, record := range testRecords {
		err := db.StoreSystemStats(record)
		if err != nil {
			test.Fatalf("Failed to store test record: '%v'.", err)
		}
	}

	for i, testCase := range testCases {
		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.query), &fields)

		response := core.SendTestAPIRequestFull(test, `stats/system/query`, fields, nil, testCase.email)
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

var testRecords []*stats.SystemMetrics = []*stats.SystemMetrics{
	{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(100),
		},
		CPUPercent:       1,
		MemPercent:       1,
		NetBytesSent:     1,
		NetBytesReceived: 1,
	},
	{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(200),
		},
		CPUPercent:       2,
		MemPercent:       2,
		NetBytesSent:     2,
		NetBytesReceived: 2,
	},
	{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(300),
		},
		CPUPercent:       3,
		MemPercent:       3,
		NetBytesSent:     3,
		NetBytesReceived: 3,
	},
}
