package stats

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
		query           *stats.Query
		expectedValues  []int
	}{
		// System Stats Base
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeSystemCPU,
			},
			expectedValues: []int{401},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeSystemMemory,
			},
			expectedValues: []int{402},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeSystemNetworkIn,
			},
			expectedValues: []int{403},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeSystemNetworkOut,
			},
			expectedValues: []int{404},
		},

		// API Request Stats Base
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.MetricTypeAPIRequest,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Grading Time Stats Base
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.MetricTypeGradingTime,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Task Time Stats Base
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.MetricTypeTaskTime,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Code Analysis Time Stats Base
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.MetricTypeCodeAnalysisTime,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// API Request Stats, Course Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeAssignmentID: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeAssignmentID: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeUserEmail: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeUserEmail: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "zzz",
				},
			},
		},

		// Grading Time Stats, Course Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeAssignmentID: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeAssignmentID: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeUserEmail: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeUserEmail: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "zzz",
				},
			},
		},

		// Task Time Stats, Course Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeAssignmentID: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeAssignmentID: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeUserEmail: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeUserEmail: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "zzz",
				},
			},
		},

		// Code Analysis Time Stats, Course Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeAssignmentID: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeAssignmentID: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeUserEmail: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeUserEmail: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeCourseID: "zzz",
				},
			},
		},

		// API Request Stats, Endpoint Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeEndpoint: "E1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeEndpoint: "zzz",
				},
			},
		},

		// API Request Stats, Sender Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeSender: "1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeSender: "zzz",
				},
			},
		},

		// API Request Stats, Locator Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeLocator: "11",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
				Where: map[stats.MetricAttribute]any{
					stats.MetricAttributeLocator: "zzz",
				},
			},
		},

		// API Request Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.MetricTypeAPIRequest,
			},
		},

		// Grading Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.MetricTypeGradingTime,
			},
		},

		// Task Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.MetricTypeTaskTime,
			},
		},

		// Code Analysis Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.MetricTypeCodeAnalysisTime,
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

func TestQueryTestingMetrics(test *testing.T) {
	fields := map[string]any{
		"use-testing-data": true,
		"type":             "api-request",
	}

	response := core.SendTestAPIRequestFull(test, `stats/query`, fields, nil, "server-admin")
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent QueryResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	expected := []*stats.Metric{db.TESTING_STATS_METRICS[0]}

	if !reflect.DeepEqual(expected, responseContent.Records) {
		test.Fatalf("Unexpected records. Expected: %s, Actual: %s.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent.Records))
	}
}

var testRecords []*stats.Metric = []*stats.Metric{
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.MetricTypeAPIRequest,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeSender:       "1",
			stats.MetricAttributeEndpoint:     "E1",
			stats.MetricAttributeUserEmail:    "U1",
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeCourseID:     "C1",
			stats.MetricAttributeLocator:      "11",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.MetricTypeGradingTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C1",
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C1",
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.MetricTypeCodeAnalysisTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C1",
			stats.MetricAttributeAssignmentID: "A1",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.MetricTypeAPIRequest,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeSender:       "2",
			stats.MetricAttributeEndpoint:     "E2",
			stats.MetricAttributeUserEmail:    "U1",
			stats.MetricAttributeCourseID:     "C2",
			stats.MetricAttributeAssignmentID: "A2",
			stats.MetricAttributeLocator:      "22",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.MetricTypeGradingTime,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C2",
			stats.MetricAttributeAssignmentID: "A2",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C2",
			stats.MetricAttributeAssignmentID: "A2",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.MetricTypeCodeAnalysisTime,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C2",
			stats.MetricAttributeAssignmentID: "A2",
			stats.MetricAttributeUserEmail:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.MetricTypeAPIRequest,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeSender:       "3",
			stats.MetricAttributeEndpoint:     "E3",
			stats.MetricAttributeUserEmail:    "U3",
			stats.MetricAttributeCourseID:     "C3",
			stats.MetricAttributeAssignmentID: "A3",
			stats.MetricAttributeLocator:      "33",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.MetricTypeGradingTime,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C3",
			stats.MetricAttributeAssignmentID: "A3",
			stats.MetricAttributeUserEmail:    "U2",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C3",
			stats.MetricAttributeAssignmentID: "A3",
			stats.MetricAttributeUserEmail:    "U2",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.MetricTypeCodeAnalysisTime,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "C3",
			stats.MetricAttributeAssignmentID: "A3",
			stats.MetricAttributeUserEmail:    "U2",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(401),
		Type:      stats.MetricTypeSystemCPU,
		Value:     float64(401),
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(402),
		Type:      stats.MetricTypeSystemMemory,
		Value:     float64(402),
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(403),
		Type:      stats.MetricTypeSystemNetworkIn,
		Value:     float64(403),
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(404),
		Type:      stats.MetricTypeSystemNetworkOut,
		Value:     float64(404),
	},
}
