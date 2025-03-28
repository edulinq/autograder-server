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
				Type: stats.API_REQUEST_STATS_TYPE,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.API_REQUEST_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Grading Time Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.GRADING_TIME_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Task Time Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.TASK_TIME_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// Code Analysis Time Stats Base.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
			},
			expectedValues: []int{100, 200, 300},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Sort: 1,
			},
			expectedValues: []int{300, 200, 100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type:  stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200, 300},
		},

		// API Request Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
		},

		// Grading Time Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
		},

		// Task Time Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
		},

		// Code Analysis Time Stats, Course Specific.
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
		},

		// API Request Stats, Endpoint Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ENDPOINT_KEY: "E1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ENDPOINT_KEY: "zzz",
				},
			},
		},

		// API Request Stats, Sender Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.SENDER_KEY: "1",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.SENDER_KEY: "zzz",
				},
			},
		},

		// API Request Stats, Locator Specific
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.LOCATOR_KEY: "11",
				},
			},
			expectedValues: []int{100},
		},
		{
			email: "server-admin",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.LOCATOR_KEY: "zzz",
				},
			},
		},

		// API Request Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
			},
		},

		// Grading Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
			},
		},

		// Task Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
			},
		},

		// Code Analysis Time Stats, Error
		{
			email:           "server-user",
			expectedLocator: "-041",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
			},
		},

		// Validation Error
		{
			email:           "server-admin",
			expectedLocator: "-301",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.UNKNOWN_METRIC_ATTRIBUTE_KEY: nil,
				},
			},
		},
		{
			email:           "server-admin",
			expectedLocator: "-301",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: nil,
				},
			},
		},
		{
			email:           "server-admin",
			expectedLocator: "-301",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "",
				},
			},
		},
		{
			email:           "server-admin",
			expectedLocator: "-301",
			query:           &stats.Query{},
		},
		{
			email:           "server-admin",
			expectedLocator: "-301",
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
		Type:      stats.API_REQUEST_STATS_TYPE,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.SENDER_KEY:        "1",
			stats.ENDPOINT_KEY:      "E1",
			stats.USER_EMAIL_KEY:    "U1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.COURSE_ID_KEY:     "C1",
			stats.LOCATOR_KEY:       "11",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.GRADING_TIME_STATS_TYPE,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.TASK_TIME_STATS_TYPE,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.API_REQUEST_STATS_TYPE,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.SENDER_KEY:        "2",
			stats.ENDPOINT_KEY:      "E2",
			stats.USER_EMAIL_KEY:    "U1",
			stats.COURSE_ID_KEY:     "C2",
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.LOCATOR_KEY:       "22",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.GRADING_TIME_STATS_TYPE,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C2",
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.TASK_TIME_STATS_TYPE,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C2",
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C2",
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.API_REQUEST_STATS_TYPE,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.SENDER_KEY:        "3",
			stats.ENDPOINT_KEY:      "E3",
			stats.USER_EMAIL_KEY:    "U3",
			stats.COURSE_ID_KEY:     "C3",
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.LOCATOR_KEY:       "33",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.GRADING_TIME_STATS_TYPE,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C3",
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.USER_EMAIL_KEY:    "U2",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.TASK_TIME_STATS_TYPE,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C3",
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.USER_EMAIL_KEY:    "U2",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C3",
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.USER_EMAIL_KEY:    "U2",
		},
	},
}
