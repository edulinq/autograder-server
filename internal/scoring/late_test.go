package scoring

import (
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// Ensure that the late days info struct can be serialized
// (since it is serialized in a Must function).
func TestLateDaysInfoStruct(test *testing.T) {
	testCases := []*LateDaysInfo{
		nil,
		&LateDaysInfo{},
		&LateDaysInfo{1, timestamp.Now(), map[string]int{"A": 1, "B": 2}, LATE_DAYS_STRUCT_VERSION, "foo", "bar"},
	}

	for _, testCase := range testCases {
		util.MustToJSON(testCase)
	}
}

// Ensure that the LateDaysInfo struct can be identified as an autograder comment.
func TestLateDaysInfoJSONContainsAutograderKey(test *testing.T) {
	content := util.MustToJSON(LateDaysInfo{})

	if !strings.Contains(content, common.AUTOGRADER_COMMENT_IDENTITY_KEY) {
		test.Fatalf("JSON does not contain autograder substring '%s': '%s'.",
			common.AUTOGRADER_COMMENT_IDENTITY_KEY, content)
	}
}

func TestComputeLateDays(test *testing.T) {
	var dayMSecs int64 = 24 * 60 * 60 * 1000

	testCases := []struct {
		dueDate        timestamp.Timestamp
		submissionTime timestamp.Timestamp
		expected       int
	}{
		{timestamp.Timestamp(0), timestamp.Timestamp(0), 0},
		{timestamp.Timestamp(0), timestamp.Timestamp(-1), 0},

		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp(dayMSecs), 0},
		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp(dayMSecs - 1), 0},

		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs), 1},
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs + 1), 2},
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs - 1), 1},

		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp((2 * dayMSecs) - 1), 1},
		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp((2 * dayMSecs) + 0), 1},
		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp((2 * dayMSecs) + 1), 2},
	}

	for i, testCase := range testCases {
		actual := computeLateDays(testCase.dueDate, testCase.submissionTime, 0)
		if testCase.expected != actual {
			test.Errorf("Case %d: Bad late days. Expected: %d, Actual: %d.", i, testCase.expected, actual)
		}
	}
}

func TestComputeLateDaysWithGraceTime(test *testing.T) {
	var dayMSecs int64 = 24 * 60 * 60 * 1000
	var minuteMSecs int64 = 60 * 1000

	testCases := []struct {
		dueDate        timestamp.Timestamp
		submissionTime timestamp.Timestamp
		graceMinutes   int
		expected       int
	}{
		// No late time, with grace time should still be 0.
		{timestamp.Timestamp(0), timestamp.Timestamp(0), 10, 0},

		// Submission within grace period should be 0.
		{timestamp.Timestamp(0), timestamp.Timestamp(5 * minuteMSecs), 10, 0},
		{timestamp.Timestamp(0), timestamp.Timestamp(10 * minuteMSecs), 10, 0},

		// Submission just after grace period should be 1 day late.
		{timestamp.Timestamp(0), timestamp.Timestamp((10 * minuteMSecs) + 1), 10, 1},

		// Submission 1 day + grace time should be 1 day late.
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs + (10 * minuteMSecs)), 10, 1},

		// Submission 1 day + grace time + 1ms should be 2 days late.
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs + (10 * minuteMSecs) + 1), 10, 2},

		// Zero grace time should behave like original function.
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs), 0, 1},
		{timestamp.Timestamp(0), timestamp.Timestamp(1), 0, 1},
	}

	for i, testCase := range testCases {
		actual := computeLateDays(testCase.dueDate, testCase.submissionTime, testCase.graceMinutes)
		if testCase.expected != actual {
			test.Errorf("Case %d: Bad late days with grace time. Expected: %d, Actual: %d.", i, testCase.expected, actual)
		}
	}
}
