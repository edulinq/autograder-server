package scoring

import (
    "strings"
    "testing"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/util"
)

// Ensure that the late days info struct can be serialized
// (since it is serialized in a Must function).
func TestLateDaysInfoStruct(test *testing.T) {
    testCases := []*LateDaysInfo{
        nil,
        &LateDaysInfo{},
        &LateDaysInfo{1, common.NowTimestamp(), map[string]int{"A": 1, "B": 2}, LATE_DAYS_STRUCT_VERSION, "foo", "bar"},
    };

    for _, testCase := range testCases {
        util.MustToJSON(testCase);
    }
}

// Ensure that the LateDaysInfo struct can be identified as an autograder comment.
func TestLateDaysInfoJSONContainsAutograderKey(test *testing.T) {
    content := util.MustToJSON(LateDaysInfo{});

    if (!strings.Contains(content, common.AUTOGRADER_COMMENT_IDENTITY_KEY)) {
        test.Fatalf("JSON does not contain autograder substring '%s': '%s'.",
                common.AUTOGRADER_COMMENT_IDENTITY_KEY, content);
    }
}
