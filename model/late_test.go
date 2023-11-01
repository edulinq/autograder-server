package model

import (
    "testing"
    "time"

    "github.com/eriq-augustine/autograder/util"
)

// Ensure that the late days info struct can be serialized
// (since it is serialized in a Must function).
func TestLateDaysInfoStruct(test *testing.T) {
    testCases := []*LateDaysInfo{
        nil,
        &LateDaysInfo{},
        &LateDaysInfo{1, time.Now(), map[string]int{"A": 1, "B": 2}, 1, "foo", "bar"},
    };

    for _, testCase := range testCases {
        util.MustToJSON(testCase);
    }
}
