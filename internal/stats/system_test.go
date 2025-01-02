package stats

import (
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/util"
)

func TestGetSystemMetricsBase(test *testing.T) {
	testCases := []struct {
		interval int
	}{
		{0},
		{-1},
		{1},
	}

	for i, testCase := range testCases {
		results, err := GetSystemMetrics(testCase.interval)
		if err != nil {
			test.Errorf("Case %d: Failed to get system metrics: '%v'.", i, err)
			continue
		}

		// Just ensure that the result is not zero.
		if results.Time == 0 {
			test.Errorf("Case %d: Got a zero result: '%s'.", i, util.MustToJSONIndent(results))
			continue
		}
	}
}

func TestCollectSystemStats(test *testing.T) {
	defer clearBackend()

	intervalMS := 1
	waitMS := time.Duration(10)

	// Ensure that there is no backend set during testing.
	if backend != nil {
		test.Fatalf("Stats backend should not be set during testing.")
	}

	typedBackend := makeTestBackend()
	backend = typedBackend

	if len(typedBackend.system) != 0 {
		test.Fatalf("Found stored stats (%d) before collection.", len(typedBackend.system))
	}

	// Start a quick collection.
	startSystemStatsCollection(intervalMS)

	// Wait for some collection.
	time.Sleep(waitMS * time.Millisecond)

	// Stop collection.
	stopSystemStatsCollection()

	// Ensure that stats have been collected.
	count := len(typedBackend.system)
	if count == 0 {
		test.Fatalf("No system stats collected.")
	}

	// Wait again.
	time.Sleep(waitMS * time.Millisecond)

	// Ensure that no more stats have been collected.
	newCount := len(typedBackend.system)
	if count != newCount {
		test.Fatalf("Got more stats after collection.")
	}
}
