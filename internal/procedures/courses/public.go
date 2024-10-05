package courses

import (
	"errors"
	"fmt"
	"slices"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// Update a course from it's local source directory.
// This effectifly just triggers a normal update.
func UpdateFromLocalSource(course *model.Course, options CourseUpsertOptions) (*CourseUpsertResult, error) {
	result, _, err := upsertFromConfigPath(course.GetSourceConfigPath(), options)
	return result, err
}

// Upsert any courses represented by the given filespec.
// Any error that occurs will be returned.
// If an error occurs within the context of a course,
// then it will be placed in both the course's message and joined to the returned error.
func UpsertFromFileSpec(spec *common.FileSpec, options CourseUpsertOptions) ([]CourseUpsertResult, error) {
	if spec == nil {
		return nil, fmt.Errorf("No FileSpec provided.")
	}

	err := spec.Validate()
	if err != nil {
		return nil, fmt.Errorf("Given FileSpec is not valid: '%w'.", err)
	}

	tempDir, err := util.MkDirTemp("autograder-upsert-course-source-")
	if err != nil {
		return nil, fmt.Errorf("Failed to make temp source dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	err = spec.CopyTarget(common.ShouldGetCWD(), tempDir, false)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy source: '%w'.", err)
	}

	return UpsertFromDir(tempDir, options)
}

func UpsertFromDir(baseDir string, options CourseUpsertOptions) ([]CourseUpsertResult, error) {
	configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err)
	}

	var errs error = nil
	results := make([]CourseUpsertResult, 0, len(configPaths))

	for _, configPath := range configPaths {
		result, courseID, err := UpsertFromConfigPath(configPath, options)
		if err != nil {
			result = &CourseUpsertResult{
				CourseID: courseID,
				Message:  err.Error(),
			}
		}

		errs = errors.Join(errs, err)
		results = append(results, *result)
	}

	slices.SortFunc(results, compareResults)

	return results, errs
}

func UpsertFromConfigPath(path string, options CourseUpsertOptions) (*CourseUpsertResult, string, error) {
	return upsertFromConfigPath(path, options)
}
