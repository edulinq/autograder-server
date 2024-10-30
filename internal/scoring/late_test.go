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
		actual := computeLateDays(testCase.dueDate, testCase.submissionTime)
		if testCase.expected != actual {
			test.Errorf("Case %d: Bad late days. Expected: %d, Actual: %d.", i, testCase.expected, actual)
		}
	}
}
