package stats

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func runStoreCourseMetricTest(test *testing.T, storeFunc func(), expectedMetric *CourseMetric) {
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

	storeFunc()

	// Ensure that stats have been collected.
	count := len(typedBackend.course)
	if count != 1 {
		test.Fatalf("Got an unexpected number of metrics. Expected: 1, Actual: %d.", len(typedBackend.course))
	}

	// Compare the stored metric with the expected one.
	if !reflect.DeepEqual(expectedMetric, typedBackend.course[0]) {
		test.Fatalf("Stored metric is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expectedMetric), util.MustToJSONIndent(typedBackend.course[0]))
	}
}

func TestStoreCourseGradingTimeBase(test *testing.T) {
	expected := &CourseMetric{
		BaseMetric: BaseMetric{
			Timestamp: timestamp.Zero(),
		},
		CourseAssignmentEmailMetric: CourseAssignmentEmailMetric{
			CourseID:     "C",
			AssignmentID: "A",
			UserEmail:    "U",
		},
		Type:  CourseMetricTypeGradingTime,
		Value: 100,
	}

	runStoreCourseMetricTest(
		test,
		func() { AsyncStoreCourseGradingTime(timestamp.Zero(), timestamp.FromMSecs(100), "C", "A", "U") },
		expected,
	)
}

func TestStoreCourseTaskTimeBase(test *testing.T) {
	expected := &CourseMetric{
		BaseMetric: BaseMetric{
			Timestamp:  timestamp.Zero(),
			Attributes: map[string]any{ATTRIBUTE_KEY_TASK: "T"},
		},
		CourseAssignmentEmailMetric: CourseAssignmentEmailMetric{
			CourseID:     "C",
			AssignmentID: "A",
			UserEmail:    "U",
		},
		Type:  CourseMetricTypeTaskTime,
		Value: 100,
	}

	runStoreCourseMetricTest(
		test,
		func() { AsyncStoreCourseTaskTime(timestamp.Zero(), timestamp.FromMSecs(100), "C", "A", "U", "T") },
		expected,
	)
}
