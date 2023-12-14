package docker

import (
    "testing"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/util"
)

// Ensure that the image info struct can be serialized
// (since it is serialized in a Must function).
func TestImageInfoStruct(test *testing.T) {
    testCases := []*ImageInfo{
        nil,
        &ImageInfo{},
        &ImageInfo{
            "foo",
            nil, []string{},
            []string{"a"}, []*common.FileSpec{},
            nil, [][]string{},
            [][]string{[]string{"a"}, []string{"b", "c"}},
            "foo", "bar",
        },
    };

    for _, testCase := range testCases {
        util.MustToJSON(testCase);
    }
}
