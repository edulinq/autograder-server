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
		query           stats.Query
		expectedValues  []int
	}{
		// API Request Stats Base.
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
			},
			[]int{100, 200, 300},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Sort: 1,
			},
			[]int{300, 200, 100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type:  stats.API_REQUEST_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			[]int{200, 300},
		},

		// Grading Time Stats Base.
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
			},
			[]int{100, 200, 300},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Sort: 1,
			},
			[]int{300, 200, 100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type:  stats.GRADING_TIME_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			[]int{200, 300},
		},

		// Task Time Stats Base.
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
			},
			[]int{100, 200, 300},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Sort: 1,
			},
			[]int{300, 200, 100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type:  stats.TASK_TIME_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			[]int{200, 300},
		},

		// Code Analysis Time Stats Base.
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
			},
			[]int{100, 200, 300},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Sort: 1,
			},
			[]int{300, 200, 100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type:  stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			[]int{200, 300},
		},

		// API Request Stats, Course Specific.
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			[]int{200},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			[]int{100, 200},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			[]int{100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
			nil,
		},

		// Grading Time Stats, Course Specific.
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			[]int{200},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			[]int{100, 200},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			[]int{100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
			nil,
		},

		// Task Time Stats, Course Specific.
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			[]int{200},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			[]int{100, 200},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			[]int{100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
			nil,
		},

		// Code Analysis Time Stats, Course Specific.
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			[]int{200},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			[]int{100, 200},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			[]int{100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
			nil,
		},

		// API Request Stats, Endpoint Specific
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ENDPOINT_KEY: "E1",
				},
			},
			[]int{100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ENDPOINT_KEY: "zzz",
				},
			},
			nil,
		},

		// API Request Stats, Sender Specific
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.SENDER_KEY: "1",
				},
			},
			[]int{100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.SENDER_KEY: "zzz",
				},
			},
			nil,
		},

		// API Request Stats, Locator Specific
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.LOCATOR_KEY: "11",
				},
			},
			[]int{100},
		},
		{
			"server-admin",
			"",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.LOCATOR_KEY: "zzz",
				},
			},
			nil,
		},

		// API Request Stats, Error
		{
			"server-user",
			"-041",
			stats.Query{
				Type: stats.API_REQUEST_STATS_TYPE,
			},
			nil,
		},

		// Grading Time Stats, Error
		{
			"server-user",
			"-041",
			stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
			},
			nil,
		},

		// Task Time Stats, Error
		{
			"server-user",
			"-041",
			stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
			},
			nil,
		},

		// Code Analysis Time Stats, Error
		{
			"server-user",
			"-041",
			stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
			},
			nil,
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
		Attributes: map[stats.MetricAttribute]any{
			stats.SENDER_KEY:        "1",
			stats.ENDPOINT_KEY:      "E1",
			stats.USER_EMAIL_KEY:    "U1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.COURSE_ID_KEY:     "C1",
			stats.LOCATOR_KEY:       "11",
			stats.DURATION_KEY:      float64(100),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.GRADING_TIME_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
			stats.DURATION_KEY:      float64(100),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.TASK_TIME_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
			stats.DURATION_KEY:      float64(100),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
			stats.VALUE_KEY:         float64(100),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.API_REQUEST_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.SENDER_KEY:        "2",
			stats.ENDPOINT_KEY:      "E2",
			stats.USER_EMAIL_KEY:    "U1",
			stats.COURSE_ID_KEY:     "C2",
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.LOCATOR_KEY:       "22",
			stats.DURATION_KEY:      float64(200),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.GRADING_TIME_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C2",
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
			stats.DURATION_KEY:      float64(200),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.TASK_TIME_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C2",
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
			stats.DURATION_KEY:      float64(200),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C2",
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
			stats.VALUE_KEY:         float64(200),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.API_REQUEST_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.SENDER_KEY:        "3",
			stats.ENDPOINT_KEY:      "E3",
			stats.USER_EMAIL_KEY:    "U3",
			stats.COURSE_ID_KEY:     "C3",
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.LOCATOR_KEY:       "33",
			stats.DURATION_KEY:      float64(300),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.GRADING_TIME_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C3",
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.USER_EMAIL_KEY:    "U2",
			stats.DURATION_KEY:      float64(300),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.TASK_TIME_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C3",
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.USER_EMAIL_KEY:    "U2",
			stats.DURATION_KEY:      float64(300),
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C3",
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.USER_EMAIL_KEY:    "U2",
			stats.VALUE_KEY:         float64(300),
		},
	},
}
