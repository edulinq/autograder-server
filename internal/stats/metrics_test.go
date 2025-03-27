package stats

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestStoreMetrics(test *testing.T) {
	defer clearBackend()

	metrics := []BaseMetric{
		// API Request.
		{
			Timestamp: timestamp.Zero(),
			Type:      API_REQUEST_STATS_KEY,
			Attributes: map[string]any{
				SENDER_KEY:        "1",
				ENDPOINT_KEY:      "E",
				USER_EMAIL_KEY:    "U",
				COURSE_ID_KEY:     "C",
				ASSIGNMENT_ID_KEY: "A",
				LOCATOR_KEY:       "2",
				DURATION_KEY:      float64(100),
			},
		},

		// Grading Time.
		{
			Timestamp: timestamp.Zero(),
			Type:      string(GRADING_TIME_STATS_KEY),
			Attributes: map[string]any{
				COURSE_ID_KEY:     "C",
				ASSIGNMENT_ID_KEY: "A",
				USER_EMAIL_KEY:    "U",
				VALUE_KEY:         float64(100),
			},
		},

		// Task Time.
		{
			Timestamp: timestamp.Zero(),
			Type:      TASK_TIME_STATS_KEY,
			Attributes: map[string]any{
				ATTRIBUTE_KEY_TASK: "T",
				COURSE_ID_KEY:      "C",
				ASSIGNMENT_ID_KEY:  "A",
				USER_EMAIL_KEY:     "U",
				VALUE_KEY:          float64(100),
			},
		},
	}

	if backend != nil {
		test.Fatalf("Stats backend should not be set during testing.")
	}

	for i, metric := range metrics {
		clearBackend()
		typedBackend := makeTestBackend()
		backend = typedBackend

		if len(typedBackend.metric) != 0 {
			test.Errorf("Case %d: Found stored stats (%d) before collection.", i, len(typedBackend.metric))
			continue
		}

		AsyncStoreMetric(&metric)

		// Ensure that stats have been collected.
		count := len(typedBackend.metric)
		if count != 1 {
			test.Errorf("Case %d: Got an unexpected number of metrics. Expected: 1, Actual: %d.", i, len(typedBackend.metric))
			continue
		}

		// Compare the stored metric with the expected one.
		if !reflect.DeepEqual(util.MustToJSON(metric), util.MustToJSON(typedBackend.metric[0])) {
			test.Errorf("Case %d: Stored metric is not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(metric), util.MustToJSONIndent(typedBackend.metric[0]))
		}
	}

}
