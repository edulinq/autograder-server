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
		email            string
		permErrorLocator string
		query            stats.MetricQuery
		expectedValues   []int
	}{
		// Grading Time Stats Base.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.GRADING_TIME_STATS_KEY,
				},
			},
			[]int{100, 200, 300},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.GRADING_TIME_STATS_KEY,
					Sort: 1,
				},
			},
			[]int{300, 200, 100},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type:  stats.GRADING_TIME_STATS_KEY,
					After: timestamp.FromMSecs(150),
				},
			},
			[]int{200, 300},
		},

		// Task Time Stats Base.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.TASK_TIME_STATS_KEY,
				},
			},
			[]int{100, 200, 300},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.TASK_TIME_STATS_KEY,
					Sort: 1,
				},
			},
			[]int{300, 200, 100},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type:  stats.TASK_TIME_STATS_KEY,
					After: timestamp.FromMSecs(150),
				},
			},
			[]int{200, 300},
		},

		// Grading Time Stats, Course Specific.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.GRADING_TIME_STATS_KEY,
					Where: map[string]any{
						stats.ASSIGNMENT_ID_KEY: "A2",
					},
				},
			},
			[]int{200},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.GRADING_TIME_STATS_KEY,
					Where: map[string]any{
						stats.ASSIGNMENT_ID_KEY: "zzz",
					},
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.GRADING_TIME_STATS_KEY,
					Where: map[string]any{
						stats.USER_EMAIL_KEY: "U1",
					},
				},
			},
			[]int{100, 200},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.GRADING_TIME_STATS_KEY,
					Where: map[string]any{
						stats.USER_EMAIL_KEY: "zzz",
					},
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.GRADING_TIME_STATS_KEY,
					Where: map[string]any{
						stats.COURSE_ID_KEY: db.TEST_COURSE_ID,
					},
				},
			},
			[]int{100, 200, 300},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.GRADING_TIME_STATS_KEY,
					Where: map[string]any{
						stats.COURSE_ID_KEY: "zzz",
					},
				},
			},
			[]int{100, 200, 300},
		},

		// Task Time Stats, Course Specific.
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.TASK_TIME_STATS_KEY,
					Where: map[string]any{
						stats.ASSIGNMENT_ID_KEY: "A2",
					},
				},
			},
			[]int{200},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.TASK_TIME_STATS_KEY,
					Where: map[string]any{
						stats.ASSIGNMENT_ID_KEY: "zzz",
					},
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.TASK_TIME_STATS_KEY,
					Where: map[string]any{
						stats.USER_EMAIL_KEY: "U1",
					},
				},
			},
			[]int{100, 200},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.TASK_TIME_STATS_KEY,
					Where: map[string]any{
						stats.USER_EMAIL_KEY: "zzz",
					},
				},
			},
			nil,
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.TASK_TIME_STATS_KEY,
					Where: map[string]any{
						stats.COURSE_ID_KEY: db.TEST_COURSE_ID,
					},
				},
			},
			[]int{100, 200, 300},
		},
		{
			"server-admin",
			"",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.TASK_TIME_STATS_KEY,
					Where: map[string]any{
						stats.COURSE_ID_KEY: "zzz",
					},
				},
			},
			[]int{100, 200, 300},
		},

		// Grading Time Stats, Error
		{
			"server-user",
			"-040",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.GRADING_TIME_STATS_KEY,
				},
			},
			nil,
		},

		// Task Time Stats, Error
		{
			"server-user",
			"-040",
			stats.MetricQuery{
				BaseQuery: stats.BaseQuery{
					Type: stats.TASK_TIME_STATS_KEY,
				},
			},
			nil,
		},
	}

	for _, record := range testRecords {
		err := db.StoreMetric(record)
		if err != nil {
			test.Fatalf("Failed to store test record: '%v'.", err)
		}
	}

	for i, testCase := range testCases {
		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.query), &fields)

		response := core.SendTestAPIRequestFull(test, `courses/stats/query`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.permErrorLocator != "" {
				if testCase.permErrorLocator != response.Locator {
					test.Errorf("Case %d: Incorrect locator on perm error. Expected: '%s', Actual: '%s'.", i, testCase.permErrorLocator, response.Locator)
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

var testRecords []*stats.BaseMetric = []*stats.BaseMetric{
	&stats.BaseMetric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.GRADING_TIME_STATS_KEY,
		Attributes: map[string]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
			stats.VALUE_KEY:         float64(100),
		},
	},
	&stats.BaseMetric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.TASK_TIME_STATS_KEY,
		Attributes: map[string]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
			stats.VALUE_KEY:         float64(100),
		},
	},
	&stats.BaseMetric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.GRADING_TIME_STATS_KEY,
		Attributes: map[string]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
			stats.VALUE_KEY:         float64(200),
		},
	},
	&stats.BaseMetric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.TASK_TIME_STATS_KEY,
		Attributes: map[string]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
			stats.VALUE_KEY:         float64(200),
		},
	},
	&stats.BaseMetric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.GRADING_TIME_STATS_KEY,
		Attributes: map[string]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.USER_EMAIL_KEY:    "U2",
			stats.VALUE_KEY:         float64(300),
		},
	},
	&stats.BaseMetric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.TASK_TIME_STATS_KEY,
		Attributes: map[string]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A3",
			stats.USER_EMAIL_KEY:    "U2",
			stats.VALUE_KEY:         float64(300),
		},
	},
}
