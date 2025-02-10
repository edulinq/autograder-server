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
			PostStaticFileOperations: []*common.FileOperation{},
			PostSubmissionFileOperations: []*common.FileOperation{
				common.NewFileOperation([]string{"a"}),
				common.NewFileOperation([]string{"b", "c"}),
			},
			Name: "foo",
			BaseDirFunc: func() string {
				return "bar"
			},
		},
	}

	for _, testCase := range testCases {
		util.MustToJSON(testCase)
	}
}
