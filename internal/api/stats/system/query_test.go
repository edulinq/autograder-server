package system

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
		email          string
		permError      bool
		query          stats.BaseQuery
		expectedValues []map[string]any
	}{
		{"server-admin", false, stats.BaseQuery{}, []map[string]any{
			{"cpu-percent": 1, "mem-percent": 1, "net-bytes-received": 1, "net-bytes-sent": 1, "timestamp": 100},
			{"cpu-percent": 2, "mem-percent": 2, "net-bytes-received": 2, "net-bytes-sent": 2, "timestamp": 200},
			{"cpu-percent": 3, "mem-percent": 3, "net-bytes-received": 3, "net-bytes-sent": 3, "timestamp": 300},
		}},
		{"server-admin", false, stats.BaseQuery{Sort: 1}, []map[string]any{
			{"cpu-percent": 3, "mem-percent": 3, "net-bytes-received": 3, "net-bytes-sent": 3, "timestamp": 300},
			{"cpu-percent": 2, "mem-percent": 2, "net-bytes-received": 2, "net-bytes-sent": 2, "timestamp": 200},
			{"cpu-percent": 1, "mem-percent": 1, "net-bytes-received": 1, "net-bytes-sent": 1, "timestamp": 100},
		}},
		{"server-admin", false, stats.BaseQuery{After: timestamp.FromMSecs(150)}, []map[string]any{
			{"cpu-percent": 2, "mem-percent": 2, "net-bytes-received": 2, "net-bytes-sent": 2, "timestamp": 200},
			{"cpu-percent": 3, "mem-percent": 3, "net-bytes-received": 3, "net-bytes-sent": 3, "timestamp": 300},
		}},

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

		var responseContent stats.QueryResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		actualSlice := make([]any, len(responseContent.Response))
		for i, data := range responseContent.Response {
			actualSlice[i] = data
		}

		expectedSlice := make([]any, len(testCase.expectedValues))
		for i, data := range testCase.expectedValues {
			expectedSlice[i] = data
		}

		expected := util.MustToGenericJSON(actualSlice, stats.SortFunc)
		actual := util.MustToGenericJSON(expectedSlice, stats.SortFunc)

		if !reflect.DeepEqual(expected, actual) {
			test.Errorf("Case %d: Response is not as expected. Expected: '%v', Actual: '%v'.", i, util.MustToJSONIndent(testCase.expectedValues), util.MustToJSONIndent(responseContent.Response))
			continue
		}
	}
}

var testRecords []*stats.SystemMetrics = []*stats.SystemMetrics{
	&stats.SystemMetrics{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(100),
		},
		CPUPercent:       1,
		MemPercent:       1,
		NetBytesSent:     1,
		NetBytesReceived: 1,
	},
	&stats.SystemMetrics{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(200),
		},
		CPUPercent:       2,
		MemPercent:       2,
		NetBytesSent:     2,
		NetBytesReceived: 2,
	},
	&stats.SystemMetrics{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(300),
		},
		CPUPercent:       3,
		MemPercent:       3,
		NetBytesSent:     3,
		NetBytesReceived: 3,
	},
}
