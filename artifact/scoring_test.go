package artifact

import (
    "testing"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/util"
)

// Ensure that the scoring info struct can be serialized
// (since it is serialized in a Must function).
func TestScoringInfoStruct(test *testing.T) {
    testCases := []*ScoringInfo{
        nil,
        &ScoringInfo{},
        &ScoringInfo{"foo", common.NowTimestamp(), common.NowTimestamp(), 1.0, 2.0, false, 1, 2, true, 3, "foo", "bar"},
    };

    for _, testCase := range testCases {
        util.MustToJSON(testCase);
    }
}
