package stats

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestStoreMetrics(test *testing.T) {
	defer clearBackend()

	testCases := []struct {
		metric        *Metric
		expectedCount int
	}{
		// API Request, Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      API_REQUEST_STATS_TYPE,
				Value:     float64(100),
				Attributes: map[MetricAttribute]any{
					SENDER_KEY:        "1",
					ENDPOINT_KEY:      "E",
					USER_EMAIL_KEY:    "U",
					COURSE_ID_KEY:     "C",
					ASSIGNMENT_ID_KEY: "A",
					LOCATOR_KEY:       "2",
				},
			},
			expectedCount: 1,
		},

		// Grading Time, Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      GRADING_TIME_STATS_TYPE,
				Value:     float64(100),
				Attributes: map[MetricAttribute]any{
					COURSE_ID_KEY:     "C",
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
				},
			},
			expectedCount: 1,
		},

		// Task Time, Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      TASK_TIME_STATS_TYPE,
				Value:     float64(100),
				Attributes: map[MetricAttribute]any{
					COURSE_ID_KEY:     "C",
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
				},
			},
			expectedCount: 1,
		},

		// Code Analysis Time, Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      CODE_ANALYSIS_TIME_STATS_TYPE,
				Value:     float64(100),
				Attributes: map[MetricAttribute]any{
					COURSE_ID_KEY:     "C",
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
				},
			},
			expectedCount: 1,
		},

		// API Request, No Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      API_REQUEST_STATS_TYPE,
				Value:     float64(100),
				Attributes: map[MetricAttribute]any{
					SENDER_KEY:     "1",
					ENDPOINT_KEY:   "E",
					USER_EMAIL_KEY: "U",
					LOCATOR_KEY:    "2",
				},
			},
			expectedCount: 1,
		},

		// Grading Time, No Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      GRADING_TIME_STATS_TYPE,
				Value:     float64(100),
				Attributes: map[MetricAttribute]any{
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
				},
			},
		},

		// Task Time, No Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      TASK_TIME_STATS_TYPE,
				Value:     float64(100),
				Attributes: map[MetricAttribute]any{
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
				},
			},
		},

		// Code Analysis Time, No Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      CODE_ANALYSIS_TIME_STATS_TYPE,
				Value:     float64(100),
				Attributes: map[MetricAttribute]any{
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
				},
			},
		},

		// Validation Error.
		{
			metric: nil,
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      UNKNOWN_METRIC_ATTRIBUTE_TYPE,
			},
		},
		{
			metric: &Metric{
				Type: API_REQUEST_STATS_TYPE,
			},
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      API_REQUEST_STATS_TYPE,
				Attributes: map[MetricAttribute]any{
					UNKNOWN_METRIC_ATTRIBUTE_KEY: "",
				},
			},
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      API_REQUEST_STATS_TYPE,
				Attributes: map[MetricAttribute]any{
					COURSE_ID_KEY: nil,
				},
			},
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      API_REQUEST_STATS_TYPE,
				Attributes: map[MetricAttribute]any{
					COURSE_ID_KEY: "",
				},
			},
		},
	}

	if backend != nil {
		test.Fatalf("Stats backend should not be set during testing.")
	}

	for i, testCase := range testCases {
		clearBackend()
		typedBackend := makeTestBackend()
		backend = typedBackend

		if len(typedBackend.metric) != 0 {
			test.Errorf("Case %d: Found stored stats (%d) before collection.", i, len(typedBackend.metric))
			continue
		}

		AsyncStoreMetric(testCase.metric)

		// Ensure that stats have been collected.
		count := len(typedBackend.metric)
		if count != testCase.expectedCount {
			test.Errorf("Case %d: Got an unexpected number of metrics. Expected: %d, Actual: %d.", i, testCase.expectedCount, len(typedBackend.metric))
			continue
		}

		// Skip comparing metrics if we're not expecting any metrics.
		if testCase.expectedCount == 0 {
			continue
		}

		// Compare the stored metric with the expected one.
		if !reflect.DeepEqual(util.MustToJSON(testCase.metric), util.MustToJSON(typedBackend.metric[0])) {
			test.Errorf("Case %d: Stored metric is not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.metric), util.MustToJSONIndent(typedBackend.metric[0]))
			continue
		}
	}
}
