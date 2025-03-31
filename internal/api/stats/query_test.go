package stats

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
		query           *stats.Query
		expectedValues  []int
	}{
		// API Request Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.API_Request_Stats_Type,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Grading Time Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.Grading_Time_Stats_Type,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Task Time Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.Task_Time_Stats_Type,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Code Analysis Time Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.Code_Analysis_Time_Stats_Type,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// API Request Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Assignment_ID_Key: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Assignment_ID_Key: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.User_Email_Key: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.User_Email_Key: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "zzz",
				},
			},
		},

		// Grading Time Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Assignment_ID_Key: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Assignment_ID_Key: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.User_Email_Key: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.User_Email_Key: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "zzz",
				},
			},
		},

		// Task Time Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Assignment_ID_Key: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Assignment_ID_Key: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.User_Email_Key: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.User_Email_Key: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "zzz",
				},
			},
		},

		// Code Analysis Time Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Assignment_ID_Key: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Assignment_ID_Key: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.User_Email_Key: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.User_Email_Key: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Course_ID_Key: "zzz",
				},
			},
		},

		// API Request Stats, Endpoint Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Endpoint_Key: "E1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Endpoint_Key: "zzz",
				},
			},
		},

		// API Request Stats, Sender Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Sender_Key: "1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Sender_Key: "zzz",
				},
			},
		},

		// API Request Stats, Locator Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Locator_Key: "11",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					stats.Locator_Key: "zzz",
				},
			},
		},

		// API Request Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.API_Request_Stats_Type,
			},
		},

		// Grading Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.Grading_Time_Stats_Type,
			},
		},

		// Task Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.Task_Time_Stats_Type,
			},
		},

		// Code Analysis Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.Code_Analysis_Time_Stats_Type,
			},
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		for _, record := range testRecords {
			err := db.StoreMetric(record)
			if err != nil {
				test.Fatalf("Failed to store test record: '%v'.", err)
			}
		}

		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.query), &fields)

		response := core.SendTestAPIRequestFull(test, `stats/query`, fields, nil, testCase.email)
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

var testRecords []*stats.Metric = []*stats.Metric{
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.API_Request_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.Sender_Key:        "1",
			stats.Endpoint_Key:      "E1",
			stats.User_Email_Key:    "U1",
			stats.Assignment_ID_Key: "A1",
			stats.Course_ID_Key:     "C1",
			stats.Locator_Key:       "11",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.Grading_Time_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C1",
			stats.Assignment_ID_Key: "A1",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.Task_Time_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C1",
			stats.Assignment_ID_Key: "A1",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.Code_Analysis_Time_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C1",
			stats.Assignment_ID_Key: "A1",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.API_Request_Stats_Type,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.Sender_Key:        "2",
			stats.Endpoint_Key:      "E2",
			stats.User_Email_Key:    "U1",
			stats.Course_ID_Key:     "C2",
			stats.Assignment_ID_Key: "A2",
			stats.Locator_Key:       "22",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.Grading_Time_Stats_Type,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C2",
			stats.Assignment_ID_Key: "A2",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.Task_Time_Stats_Type,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C2",
			stats.Assignment_ID_Key: "A2",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.Code_Analysis_Time_Stats_Type,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C2",
			stats.Assignment_ID_Key: "A2",
			stats.User_Email_Key:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.API_Request_Stats_Type,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.Sender_Key:        "3",
			stats.Endpoint_Key:      "E3",
			stats.User_Email_Key:    "U3",
			stats.Course_ID_Key:     "C3",
			stats.Assignment_ID_Key: "A3",
			stats.Locator_Key:       "33",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.Grading_Time_Stats_Type,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C3",
			stats.Assignment_ID_Key: "A3",
			stats.User_Email_Key:    "U2",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.Task_Time_Stats_Type,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C3",
			stats.Assignment_ID_Key: "A3",
			stats.User_Email_Key:    "U2",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.Code_Analysis_Time_Stats_Type,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.Course_ID_Key:     "C3",
			stats.Assignment_ID_Key: "A3",
			stats.User_Email_Key:    "U2",
		},
	},
}
