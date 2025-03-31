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
				Type: stats.APIRequestStatsType,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.APIRequestStatsType,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Grading Time Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GradingTimeStatsType,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GradingTimeStatsType,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.GradingTimeStatsType,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Task Time Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TaskTimeStatsType,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TaskTimeStatsType,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.TaskTimeStatsType,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Code Analysis Time Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CodeAnalysisTimeStatsType,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CodeAnalysisTimeStatsType,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.CodeAnalysisTimeStatsType,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// API Request Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.AssignmentIDKey: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.AssignmentIDKey: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.UserEmailKey: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.UserEmailKey: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.CourseIDKey: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.CourseIDKey: "zzz",
				},
			},
		},

		// Grading Time Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GradingTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.AssignmentIDKey: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GradingTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.AssignmentIDKey: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GradingTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.UserEmailKey: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GradingTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.UserEmailKey: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GradingTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.CourseIDKey: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GradingTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.CourseIDKey: "zzz",
				},
			},
		},

		// Task Time Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TaskTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.AssignmentIDKey: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TaskTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.AssignmentIDKey: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TaskTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.UserEmailKey: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TaskTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.UserEmailKey: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TaskTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.CourseIDKey: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TaskTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.CourseIDKey: "zzz",
				},
			},
		},

		// Code Analysis Time Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CodeAnalysisTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.AssignmentIDKey: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CodeAnalysisTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.AssignmentIDKey: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CodeAnalysisTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.UserEmailKey: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CodeAnalysisTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.UserEmailKey: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CodeAnalysisTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.CourseIDKey: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CodeAnalysisTimeStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.CourseIDKey: "zzz",
				},
			},
		},

		// API Request Stats, Endpoint Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.EndpointKey: "E1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.EndpointKey: "zzz",
				},
			},
		},

		// API Request Stats, Sender Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.SenderKey: "1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.SenderKey: "zzz",
				},
			},
		},

		// API Request Stats, Locator Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.LocatorKey: "11",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
				Where: map[stats.MetricAttribute]any{
					stats.LocatorKey: "zzz",
				},
			},
		},

		// API Request Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.APIRequestStatsType,
			},
		},

		// Grading Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.GradingTimeStatsType,
			},
		},

		// Task Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.TaskTimeStatsType,
			},
		},

		// Code Analysis Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.CodeAnalysisTimeStatsType,
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
		Type:      stats.APIRequestStatsType,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.SenderKey:       "1",
			stats.EndpointKey:     "E1",
			stats.UserEmailKey:    "U1",
			stats.AssignmentIDKey: "A1",
			stats.CourseIDKey:     "C1",
			stats.LocatorKey:      "11",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.GradingTimeStatsType,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.CourseIDKey:     "C1",
			stats.AssignmentIDKey: "A1",
			stats.UserEmailKey:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.TaskTimeStatsType,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.CourseIDKey:     "C1",
			stats.AssignmentIDKey: "A1",
			stats.UserEmailKey:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.CodeAnalysisTimeStatsType,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.CourseIDKey:     "C1",
			stats.AssignmentIDKey: "A1",
			stats.UserEmailKey:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.APIRequestStatsType,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.SenderKey:       "2",
			stats.EndpointKey:     "E2",
			stats.UserEmailKey:    "U1",
			stats.CourseIDKey:     "C2",
			stats.AssignmentIDKey: "A2",
			stats.LocatorKey:      "22",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.GradingTimeStatsType,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.CourseIDKey:     "C2",
			stats.AssignmentIDKey: "A2",
			stats.UserEmailKey:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.TaskTimeStatsType,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.CourseIDKey:     "C2",
			stats.AssignmentIDKey: "A2",
			stats.UserEmailKey:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.CodeAnalysisTimeStatsType,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.CourseIDKey:     "C2",
			stats.AssignmentIDKey: "A2",
			stats.UserEmailKey:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.APIRequestStatsType,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.SenderKey:       "3",
			stats.EndpointKey:     "E3",
			stats.UserEmailKey:    "U3",
			stats.CourseIDKey:     "C3",
			stats.AssignmentIDKey: "A3",
			stats.LocatorKey:      "33",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.GradingTimeStatsType,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.CourseIDKey:     "C3",
			stats.AssignmentIDKey: "A3",
			stats.UserEmailKey:    "U2",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.TaskTimeStatsType,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.CourseIDKey:     "C3",
			stats.AssignmentIDKey: "A3",
			stats.UserEmailKey:    "U2",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.CodeAnalysisTimeStatsType,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.CourseIDKey:     "C3",
			stats.AssignmentIDKey: "A3",
			stats.UserEmailKey:    "U2",
		},
	},
}
