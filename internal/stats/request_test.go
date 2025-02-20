package stats

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestStoreRequestMetric(test *testing.T) {
	defer clearBackend()

	expectedMetric := &RequestMetric{
		BaseMetric: BaseMetric{
			Timestamp: timestamp.Zero(),
		},
		Endpoint:     "E",
		Locator:      "1",
		CourseID:     "C",
		AssignmentID: "A",
		UserEmail:    "U",
		Sender:       "2",
		Duration:     100,
	}

	// Ensure that there is no backend set during testing.
	if backend != nil {
		test.Fatalf("Stats backend should not be set during testing.")
	}

	typedBackend := makeTestBackend()
	backend = typedBackend

	if len(typedBackend.request) != 0 {
		test.Fatalf("Found stored stats (%d) before collection.", len(typedBackend.request))
	}

	AsyncStoreRequestMetric(timestamp.Zero(), timestamp.FromMSecs(100), "C", "A", "U", "E", "1", "2")

	// Ensure that stats have been collected.
	count := len(typedBackend.request)
	if count != 1 {
		test.Fatalf("Got an unexpected number of metrics. Expected: 1, Actual: %d.", len(typedBackend.request))
	}

	// Compare the stored metric with the expected one.
	if !reflect.DeepEqual(expectedMetric, typedBackend.request[0]) {
		test.Fatalf("Stored metric is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expectedMetric), util.MustToJSONIndent(typedBackend.request[0]))
	}

}
