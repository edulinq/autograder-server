package stats

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestStoreAPIRequestMetric(test *testing.T) {
	defer clearBackend()

	expectedMetric := &BaseMetric{
		Timestamp: timestamp.Zero(),
		Attributes: map[string]any{
			SENDER:        "1",
			ENDPOINT:      "E",
			USER_EMAIL:    "U",
			COURSE_ID:     "C",
			ASSIGNMENT_ID: "A",
			LOCATOR:       "2",
			DURATION:      uint64(100),
		},
	}

	// Ensure that there is no backend set during testing.
	if backend != nil {
		test.Fatalf("Stats backend should not be set during testing.")
	}

	typedBackend := makeTestBackend()
	backend = typedBackend

	if len(typedBackend.apiRequest) != 0 {
		test.Fatalf("Found stored stats (%d) before collection.", len(typedBackend.apiRequest))
	}

	AsyncStoreAPIRequestMetric(timestamp.Zero(), timestamp.FromMSecs(100), "1", "E", "U", "C", "A", "2")

	// Ensure that stats have been collected.
	count := len(typedBackend.apiRequest)
	if count != 1 {
		test.Fatalf("Got an unexpected number of metrics. Expected: 1, Actual: %d.", len(typedBackend.apiRequest))
	}

	// Compare the stored metric with the expected one.
	if !reflect.DeepEqual(expectedMetric, typedBackend.apiRequest[0]) {
		test.Fatalf("Stored metric is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expectedMetric), util.MustToJSONIndent(typedBackend.apiRequest[0]))
	}
}
