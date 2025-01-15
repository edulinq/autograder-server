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

	err := storeCourseGradingTime(timestamp.Zero(), timestamp.FromMSecs(100), "C", "A", "U")
	if err != nil {
		test.Fatalf("Failed to store grading time: '%v'.", err)
	}

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
