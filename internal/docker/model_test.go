package docker

import (
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"
)

// Ensure that the image info struct can be serialized
// (since it is serialized in a Must function).
func TestImageInfoStruct(test *testing.T) {
	testCases := []*ImageInfo{
		nil,
		&ImageInfo{},
		&ImageInfo{
			Image:                    "foo",
			PreStaticDockerCommands:  nil,
			PostStaticDockerCommands: []string{},
			Invocation:               []string{"a"},
			StaticFiles:              []*common.FileSpec{},
			PreStaticFileOperations:  nil,
			PostStaticFileOperations: []common.FileOperation{},
			PostSubmissionFileOperations: []common.FileOperation{
				common.FileOperation([]string{"a"}),
				common.FileOperation([]string{"b", "c"}),
			},
			Name:    "foo",
			BaseDir: "bar",
		},
	}

	for _, testCase := range testCases {
		util.MustToJSON(testCase)
	}
}
