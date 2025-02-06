package stats

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestStoreCourseGradingTimeBase(test *testing.T) {
	defer clearBackend()

	// Ensure that there is no backend set during testing.
	if backend != nil {
		test.Fatalf("Stats backend should not be set during testing.")
	}

	typedBackend := makeTestBackend()
	backend = typedBackend

	if len(typedBackend.course) != 0 {
		test.Fatalf("Found stored stats (%d) before collection.", len(typedBackend.course))
	}

	expected := &CourseMetric{
		BaseMetric: BaseMetric{
			Timestamp: timestamp.Zero(),
		},
		Type:         CourseMetricTypeGradingTime,
		CourseID:     "C",
		AssignmentID: "A",
		UserEmail:    "U",
		Value:        100,
	}

	AsyncStoreCourseGradingTime(timestamp.Zero(), timestamp.FromMSecs(100), "C", "A", "U")

	// Ensure that stats have been collected.
	count := len(typedBackend.course)
	if count != 1 {
		test.Fatalf("Got an unexpected number of metrics. Expected: 1, Actual: %d.", len(typedBackend.course))
	}

	if !reflect.DeepEqual(expected, typedBackend.course[0]) {
		test.Fatalf("Stored metric is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(typedBackend.course[0]))
	}
}

func TestStoreCourseStatDurationBase(test *testing.T) {
	defer clearBackend()

	// Ensure that there is no backend set during testing.
	if backend != nil {
		test.Fatalf("Stats backend should not be set during testing.")
	}

	typedBackend := makeTestBackend()
	backend = typedBackend

	if len(typedBackend.course) != 0 {
		test.Fatalf("Found stored stats (%d) before collection.", len(typedBackend.course))
	}

	expected := &CourseMetric{
		BaseMetric: BaseMetric{
			Timestamp:  timestamp.Zero(),
			Attributes: map[string]any{TASK_KEY: "T"},
		},
		Type:         CourseMetricTypeTaskDuration,
		CourseID:     "C",
		AssignmentID: "A",
		UserEmail:    "U",
		Value:        100,
	}

	AsyncStoreCourseStatDuration(timestamp.Zero(), timestamp.FromMSecs(100), "C", "A", "U", "T")

	// Ensure that stats have been collected.
	count := len(typedBackend.course)
	if count != 1 {
		test.Fatalf("Got an unexpected number of metrics. Expected: 1, Actual: %d.", len(typedBackend.course))
	}

	if !reflect.DeepEqual(expected, typedBackend.course[0]) {
		test.Fatalf("Stored metric is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(typedBackend.course[0]))
	}
}
