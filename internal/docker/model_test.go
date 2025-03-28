package docker

import (
	"testing"

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
			StaticFiles:              []*util.FileSpec{},
			PreStaticFileOperations:  nil,
			PostStaticFileOperations: []*util.FileOperation{},
			PostSubmissionFileOperations: []*util.FileOperation{
				util.NewFileOperation([]string{"a"}),
				util.NewFileOperation([]string{"b", "c"}),
			},
			Name: "foo",
			BaseDirFunc: func() (string, string) {
				return "bar", "baz"
			},
		},
	}

	for _, testCase := range testCases {
		util.MustToJSON(testCase)
	}
}
