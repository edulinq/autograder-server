package stats

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestStoreMetrics(test *testing.T) {
	defer clearBackend()

	metric := Metric{
		Timestamp: timestamp.Zero(),
		Type:      API_REQUEST_STATS_TYPE,
		Attributes: map[MetricAttribute]any{
			SENDER_KEY:        "1",
			ENDPOINT_KEY:      "E",
			USER_EMAIL_KEY:    "U",
			COURSE_ID_KEY:     "C",
			ASSIGNMENT_ID_KEY: "A",
			LOCATOR_KEY:       "2",
			DURATION_KEY:      float64(100),
		},
	}

	if backend != nil {
		test.Fatalf("Stats backend should not be set during testing.")
	}

	clearBackend()
	typedBackend := makeTestBackend()
	backend = typedBackend

	if len(typedBackend.metric) != 0 {
		test.Errorf("Found stored stats (%d) before collection.", len(typedBackend.metric))
		return
	}

	AsyncStoreMetric(&metric)

	// Ensure that stats have been collected.
	count := len(typedBackend.metric)
	if count != 1 {
		test.Errorf("Got an unexpected number of metrics. Expected: 1, Actual: %d.", len(typedBackend.metric))
		return
	}

	// Compare the stored metric with the expected one.
	if !reflect.DeepEqual(util.MustToJSON(metric), util.MustToJSON(typedBackend.metric[0])) {
		test.Errorf("Stored metric is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(metric), util.MustToJSONIndent(typedBackend.metric[0]))
	}
}

func TestStoreCourseMetrics(test *testing.T) {
	defer clearBackend()

	testCases := []struct {
		metric        *Metric
		expectedCount int
	}{
		// Grading Time, Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.Zero(),
				Type:      GRADING_TIME_STATS_TYPE,
				Attributes: map[MetricAttribute]any{
					COURSE_ID_KEY:     "C",
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
					DURATION_KEY:      float64(100),
				},
			},
			expectedCount: 1,
		},

		// Task Time, Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.Zero(),
				Type:      TASK_TIME_STATS_TYPE,
				Attributes: map[MetricAttribute]any{
					COURSE_ID_KEY:     "C",
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
					DURATION_KEY:      float64(100),
				},
			},
			expectedCount: 1,
		},

		// Code Analysis Time, Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.Zero(),
				Type:      CODE_ANALYSIS_TIME_STATS_TYPE,
				Attributes: map[MetricAttribute]any{
					COURSE_ID_KEY:     "C",
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
					DURATION_KEY:      float64(100),
				},
			},
			expectedCount: 1,
		},

		// Grading Time, No Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.Zero(),
				Type:      GRADING_TIME_STATS_TYPE,
				Attributes: map[MetricAttribute]any{
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
					DURATION_KEY:      float64(100),
				},
			},
		},

		// Task Time, No Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.Zero(),
				Type:      TASK_TIME_STATS_TYPE,
				Attributes: map[MetricAttribute]any{
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
					DURATION_KEY:      float64(100),
				},
			},
		},

		// Code Analysis Time, No Course ID.
		{
			metric: &Metric{
				Timestamp: timestamp.Zero(),
				Type:      CODE_ANALYSIS_TIME_STATS_TYPE,
				Attributes: map[MetricAttribute]any{
					ASSIGNMENT_ID_KEY: "A",
					USER_EMAIL_KEY:    "U",
					DURATION_KEY:      float64(100),
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

		AsyncStoreCourseMetric(testCase.metric)

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
		}
	}
}
