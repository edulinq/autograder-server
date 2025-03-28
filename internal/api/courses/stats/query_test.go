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
		// Grading Time Stats Base.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Sort: 1,
			},
			expectedValues: []int{200, 100},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type:  stats.GRADING_TIME_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200},
		},

		// Task Time Stats Base.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Sort: 1,
			},
			expectedValues: []int{200, 100},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type:  stats.TASK_TIME_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200},
		},

		// Code Analysis Time Stats Base.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Sort: 1,
			},
			expectedValues: []int{200, 100},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type:  stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				After: timestamp.FromMSecs(150),
			},
			expectedValues: []int{200},
		},

		// Grading Time Stats, Course Specific.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: db.TEST_COURSE_ID,
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			expectedValues: []int{100, 200},
		},

		// Task Time Stats, Course Specific.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: db.TEST_COURSE_ID,
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			expectedValues: []int{100, 200},
		},

		// Code Analysis Time Stats, Course Specific.
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "A2",
				},
			},
			expectedValues: []int{200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.ASSIGNMENT_ID_KEY: "zzz",
				},
			},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "U1",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.USER_EMAIL_KEY: "zzz",
				},
			},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: db.TEST_COURSE_ID,
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "zzz",
				},
			},
			expectedValues: []int{100, 200},
		},
		{
			email: "course-admin",
			query: &stats.Query{
				Type: stats.CODE_ANALYSIS_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "C1",
				},
			},
			expectedValues: []int{100, 200},
		},

		// Grading Time Stats, Error
		{
			email:           "course-user",
			expectedLocator: "-017",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
			},
		},

		// Task Time Stats, Error
		{
			email:           "course-user",
			expectedLocator: "-017",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
			},
		},

		// Code Analysis Time Stats, Error
		{
			email:           "course-user",
			expectedLocator: "-017",
			query: &stats.Query{
				Type: stats.TASK_TIME_STATS_TYPE,
			},
		},

		// Validation Error.
		{
			email:           "course-admin",
			expectedLocator: "-618",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.UNKNOWN_METRIC_ATTRIBUTE_KEY: nil,
				},
			},
		},
		{
			email:           "course-admin",
			expectedLocator: "-618",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: nil,
				},
			},
		},
		{
			email:           "course-admin",
			expectedLocator: "-618",
			query: &stats.Query{
				Type: stats.GRADING_TIME_STATS_TYPE,
				Where: map[stats.MetricAttribute]any{
					stats.COURSE_ID_KEY: "",
				},
			},
		},
		{
			email:           "course-admin",
			expectedLocator: "-618",
			query:           &stats.Query{},
		},
		{
			email:           "course-admin",
			expectedLocator: "-618",
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
	// Context course metrics.
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.GRADING_TIME_STATS_TYPE,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.TASK_TIME_STATS_TYPE,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.GRADING_TIME_STATS_TYPE,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.TASK_TIME_STATS_TYPE,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(200),
		Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     db.TEST_COURSE_ID,
			stats.ASSIGNMENT_ID_KEY: "A2",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},

	// Non-context course metrics.
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.GRADING_TIME_STATS_TYPE,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.TASK_TIME_STATS_TYPE,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.FromMSecs(300),
		Type:      stats.CODE_ANALYSIS_TIME_STATS_TYPE,
		Value:     float64(300),
		Attributes: map[stats.MetricAttribute]any{
			stats.COURSE_ID_KEY:     "C1",
			stats.ASSIGNMENT_ID_KEY: "A1",
			stats.USER_EMAIL_KEY:    "U1",
		},
	},
}
