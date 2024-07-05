package model

import (
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"
)

// Ensure that the scoring info struct can be serialized
// (since it is serialized in a Must function).
func TestScoringInfoStruct(test *testing.T) {
	testCases := []*ScoringInfo{
		nil,
		&ScoringInfo{},
		&ScoringInfo{"foo", common.NowTimestamp(), common.NowTimestamp(), 1.0, 2.0, false, 1, 2, true, SCORING_INFO_STRUCT_VERSION, "foo", "bar"},
	}

	for _, testCase := range testCases {
		util.MustToJSON(testCase)
	}
}

// Ensure that the ScoringInfo struct can be identified as an autograder comment.
func TestScoringInfoJSONContainsAutograderKey(test *testing.T) {
	content := util.MustToJSON(ScoringInfo{})

	if !strings.Contains(content, common.AUTOGRADER_COMMENT_IDENTITY_KEY) {
		test.Fatalf("JSON does not contain autograder substring '%s': '%s'.",
			common.AUTOGRADER_COMMENT_IDENTITY_KEY, content)
	}
}
