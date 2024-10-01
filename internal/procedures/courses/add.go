package courses

import (
	"fmt"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

// Add any courses represented by the given filespec.
func AddFromFileSpec(spec *common.FileSpec) ([]string, error) {
	if spec == nil {
		return []string{}, nil
	}

	err := spec.Validate()
	if err != nil {
		return nil, fmt.Errorf("Given FileSpec is not valid: '%w'.", err)
	}

	tempDir, err := util.MkDirTemp("autograder-add-course-source-")
	if err != nil {
		return nil, fmt.Errorf("Failed to make temp source dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	err = spec.CopyTarget(common.ShouldGetCWD(), tempDir, false)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy source: '%w'.", err)
	}

	courseIDs, err := db.AddCoursesFromDir(tempDir, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to add course dir: '%w'.", err)
	}

	return courseIDs, nil
}
