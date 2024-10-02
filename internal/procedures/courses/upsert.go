package courses

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const UNKNOWN_COURSE_ID = "<unknown>"

type CourseUpsertOptions struct {
	ContextUser *model.ServerUser `json:"-"`

	SkipCreates bool `json:"skip-creates"`
	SkipUpdates bool `json:"skip-updates"`

	DryRun bool `json:"dry-run"`
}

type CourseUpsertResult struct {
	CourseID string `json:"course-id"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`

	Created bool `json:"created"`
	Updated bool `json:"updated"`
	Skipped bool `json:"skipped"`
}

// Upsert any courses represented by the given filespec.
// Error handling is a bit more complex here,
// as there are three different ways to return errors.
// System errors will always be reported in the error return value.
// User errors (a user providing bad input) will either be returned in the string input,
// or as part of the Message filed in a CourseUpsertResult (where !Success) if there is a context course for the error.
// Note that the required permissions changes depending on if it is an insert or update.
func UpsertFromFileSpec(spec *common.FileSpec, options CourseUpsertOptions) ([]CourseUpsertResult, string, error) {
	if spec == nil {
		return []CourseUpsertResult{}, "No FileSpec provided.", nil
	}

	err := spec.Validate()
	if err != nil {
		return nil, fmt.Sprintf("Given FileSpec is not valid: '%v'.", err), nil
	}

	tempDir, err := util.MkDirTemp("autograder-upsert-course-source-")
	if err != nil {
		return nil, "", fmt.Errorf("Failed to make temp source dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	err = spec.CopyTarget(common.ShouldGetCWD(), tempDir, false)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to copy source: '%w'.", err)
	}

	return UpsertFromDir(tempDir, options)
}

func UpsertFromDir(baseDir string, options CourseUpsertOptions) ([]CourseUpsertResult, string, error) {
	configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err)
	}

	var errs error = nil
	results := make([]CourseUpsertResult, 0, len(configPaths))

	for _, configPath := range configPaths {
		result, err := UpsertFromConfigPath(configPath, options)

		errs = errors.Join(errs, err)

		if result != nil {
			results = append(results, *result)
		}
	}

	slices.SortFunc(results, compareResults)

	return results, "", errs
}

func UpsertFromConfigPath(path string, options CourseUpsertOptions) (*CourseUpsertResult, error) {
	result := &CourseUpsertResult{}

	if options.ContextUser == nil {
		return nil, fmt.Errorf("No context user provided.")
	}

	course, _, err := model.FullLoadCourseFromPath(path)
	if err != nil {
		result.CourseID = UNKNOWN_COURSE_ID
		result.Message = fmt.Sprintf("Failed to read course config: '%v'.", err)
		return result, nil
	}

	result.CourseID = course.ID

	oldCourse, err := db.GetCourse(course.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch existing course '%s': '%w'.", course.ID, err)
	}

	err = checkPermissions(course, oldCourse, options)
	if err != nil {
		result.Message = err.Error()
		return result, nil
	}

	result.Created = (oldCourse == nil) && !options.SkipCreates
	result.Updated = (oldCourse != nil) && !options.SkipUpdates
	result.Skipped = !result.Created && !result.Updated

	// Stop early.
	if options.DryRun || result.Skipped {
		result.Success = true
		return result, nil
	}

	_, err = db.LoadCourse(path)
	if err != nil {
		// Note that we have already successfully read the course config,
		// so any error here is on the system, not the user.
		return nil, fmt.Errorf("Failed to upsert course '%s': '%w'.", course.ID, err)
	}

	result.Success = true

	return result, nil
}

func checkPermissions(course *model.Course, oldCourse *model.Course, options CourseUpsertOptions) error {
	if options.ContextUser == nil {
		return fmt.Errorf("No context user provided.")
	}

	// Server creators's can both insert and update.
	if options.ContextUser.Role >= model.ServerRoleCourseCreator {
		return nil
	}

	// Insert
	if oldCourse == nil {
		if options.ContextUser.Role < model.ServerRoleCourseCreator {
			return fmt.Errorf("User does not have sufficient server-level permissions to create a course.")
		}
	}

	// Update
	// User must be at least a course admin to update the course.
	if options.ContextUser.GetCourseRole(course.ID) < model.CourseRoleAdmin {
		return fmt.Errorf("User does not have sufficient course-level permissions to update course.")
	}

	return nil
}

func compareResults(a CourseUpsertResult, b CourseUpsertResult) int {
	return strings.Compare(a.CourseID, b.CourseID)
}
