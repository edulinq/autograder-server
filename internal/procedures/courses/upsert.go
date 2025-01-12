package courses

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lms/lmssync"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/model/tasks"
	"github.com/edulinq/autograder/internal/util"
)

// The primary course upserting function.
// Returns the result (on success), the name of the course (or UNKNOWN_COURSE_ID), and an error.
func upsertFromConfigPath(path string, options CourseUpsertOptions) (*CourseUpsertResult, string, error) {
	if options.ContextUser == nil {
		return nil, UNKNOWN_COURSE_ID, fmt.Errorf("No context user provided.")
	}

	if !util.IsFile(path) {
		return nil, UNKNOWN_COURSE_ID, fmt.Errorf("Course config path does not point to a file: '%s'.", path)
	}

	result := &CourseUpsertResult{}

	// Read the course config and perform initial checks (like permission checks) to see if we can continue.
	// Note that the course ID has been modified if this is a dry run.
	course, err := readConfigForInitialChecks(path, options, result)
	if err != nil {
		return nil, UNKNOWN_COURSE_ID, fmt.Errorf("Failed to perform initial checks on course: '%w'.", err)
	}

	if course == nil {
		return result, UNKNOWN_COURSE_ID, nil
	}

	// If we are doing an update, stop any existing tasks.
	if result.Updated && !options.DryRun {
		tasks.Handler.StopCourse(course.GetID())
	}

	// Update Source Directory
	err = updateSourceDirFromConfigPath(path, course)
	if err != nil {
		return nil, result.CourseID, fmt.Errorf("Failed to update source dir: '%w'.", err)
	}

	// Sync source
	if !options.SkipSourceSync {
		newCourse, err := syncSource(course, options)
		if err != nil {
			return nil, result.CourseID, fmt.Errorf("Failed to sync course source: '%w'.", err)
		}

		// Use the new (synced) course.
		// The new course should also have a modified ID if this is a dry run.
		if newCourse != nil {
			course = newCourse
		}
	}

	// Upsert Course Into Database
	if !options.DryRun {
		err = db.SaveCourse(course)
		if err != nil {
			return nil, result.CourseID, fmt.Errorf("Failed to save course to database: '%w'.", err)
		}
	}

	// Sync LMS
	if !options.SkipLMSSync {
		err = syncLMS(course, options, result)
		if err != nil {
			return nil, result.CourseID, fmt.Errorf("Failed to sync course with LMS: '%w'.", err)
		}
	}

	// Build Images
	if !options.SkipBuildImages {
		builtImages, err := course.BuildAssignmentImagesDefault()
		if err != nil {
			return nil, result.CourseID, fmt.Errorf("Failed to build assignment images: '%w'.", err)
		}

		result.BuiltAssignmentImages = builtImages
	}

	// Cleanup

	// Remove source if this was a dry run.
	if options.DryRun {
		err = util.RemoveDirent(course.GetBaseSourceDir())
		if err != nil {
			return nil, result.CourseID, fmt.Errorf("Failed to remove dry run source dir: '%w'.", err)
		}
	}

	result.Success = true

	return result, result.CourseID, nil
}

// Sync with the LMS.
// This will properly handle dry runs.
func syncLMS(course *model.Course, options CourseUpsertOptions, result *CourseUpsertResult) error {
	// If this is a dry run, we will temporarily convert the course ID back to the original ID.
	// This is because LMS syncing will need to reference existing users to compute update (even in a dry run).
	if options.DryRun {
		course.ID = result.CourseID

		defer func() {
			course.ID = DRY_RUN_PREFIX + course.ID
		}()
	}

	lmsSyncResult, err := lmssync.SyncLMS(course, options.DryRun, !options.SkipEmails)
	if err != nil {
		return fmt.Errorf("Failed to sync course with LMS: '%w'.", err)
	}

	result.LMSSyncResult = lmsSyncResult
	return nil
}

// Sync the courses source to the canonical source directory for the source.
// If the source was updated, return the new course represented by that source.
func syncSource(course *model.Course, options CourseUpsertOptions) (*model.Course, error) {
	source := course.GetSource()
	if (source == nil) || source.IsEmpty() || source.IsNil() {
		return nil, nil
	}

	// Copy over the source to a temp dir.
	tempDir, err := util.MkDirTemp("autograder.internal.procedures.courses.upsert-source-")
	if err != nil {
		return nil, fmt.Errorf("Failed to make temp dir for source update: '%w'.", err)
	}

	err = source.CopyTarget(common.ShouldGetCWD(), tempDir, false)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy source ('%s') to temp dir: '%w'.", source, err)
	}

	// Check for multiple courses in the source.
	configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, tempDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to search for course configs in '%s': '%w'.", tempDir, err)
	}

	if len(configPaths) == 0 {
		return nil, fmt.Errorf("Did not find any course configs in course source ('%s'), should be exactly one.", source)
	}

	if len(configPaths) > 1 {
		return nil, fmt.Errorf("Found too many course configs (%d) in course source ('%s'), should be exactly one.", len(configPaths), source)
	}

	// Ensure this course can load correctly.
	_, _, err = loadCourseConfig(configPaths[0], true, options)
	if err != nil {
		return nil, fmt.Errorf("Failed to load course from source ('%s'): '%w'.", source, err)
	}

	// Now that the source has been checked, copy over the source to the canonical location.
	err = updateSourceDirFromConfigPath(configPaths[0], course)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy updated source dir from source ('%s'): '%w'.", source, err)
	}

	// Load the course from the canonical source.
	course, _, err = loadCourseConfig(course.GetSourceConfigPath(), true, options)
	if err != nil {
		return nil, fmt.Errorf("Failed to load updated course from source ('%s'): '%w'.", source, err)
	}

	return course, nil
}

// Copy over contents from the given directory into the course's canonical source dir.
func updateSourceDirFromConfigPath(configPath string, course *model.Course) error {
	baseDir := util.ShouldAbs(filepath.Dir(configPath))
	sourceDir := util.ShouldAbs(course.GetBaseSourceDir())

	// Check if we are updating from the same source.
	if baseDir == sourceDir {
		return nil
	}

	if util.PathExists(sourceDir) {
		err := util.RemoveDirent(sourceDir)
		if err != nil {
			return fmt.Errorf("Failed to remove source dir '%s': '%w'.", sourceDir, err)
		}
	}

	err := util.CopyDirWhole(baseDir, sourceDir)
	if err != nil {
		return fmt.Errorf("Failed to copy source dir from '%s' to '%s': '%w'.", baseDir, sourceDir, err)
	}

	return nil
}

// Read the course config for initial checks.
// Return an initial (not final) version of the course if we can continue the upsert process,
// nil otherwise (which may or may not have an error).
// This does not write anything and does not assume the config is in the source directory.
func readConfigForInitialChecks(path string, options CourseUpsertOptions, result *CourseUpsertResult) (*model.Course, error) {
	course, originalCourseID, err := loadCourseConfig(path, true, options)
	if err != nil {
		return nil, fmt.Errorf("Failed to read course config: '%w'.", err)
	}

	result.CourseID = originalCourseID

	oldCourse, err := db.GetCourse(originalCourseID)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch existing course '%s': '%w'.", originalCourseID, err)
	}

	err = checkPermissions(course, oldCourse, options)
	if err != nil {
		return nil, fmt.Errorf("Insufficient permissions: '%s'.", err)
	}

	result.Created = (oldCourse == nil)
	result.Updated = (oldCourse != nil)

	return course, nil
}

func checkPermissions(course *model.Course, oldCourse *model.Course, options CourseUpsertOptions) error {
	if options.ContextUser == nil {
		return fmt.Errorf("No context user provided.")
	}

	// Server creators can both insert and update.
	if options.ContextUser.Role >= model.ServerRoleCourseCreator {
		return nil
	}

	// Insert
	if oldCourse == nil {
		if options.ContextUser.Role < model.ServerRoleCourseCreator {
			return fmt.Errorf("User does not have sufficient server-level permissions to create a course.")
		}

		return nil
	}

	// Update
	// User must be at least a course admin to update the course.
	// Remember that on a dry run the course ID will be modified, use the old course ID.
	if options.ContextUser.GetCourseRole(oldCourse.ID) < model.CourseRoleAdmin {
		return fmt.Errorf("User does not have sufficient course-level permissions to update course.")
	}

	return nil
}

// Load a course from a course config path and return the course and original ID.
// If this is a dry run, modify the course's ID.
func loadCourseConfig(path string, isSource bool, options CourseUpsertOptions) (*model.Course, string, error) {
	course, _, err := model.FullLoadCourseFromPath(path, isSource)
	if err != nil {
		return nil, "", err
	}

	originalID := course.GetID()

	// If this is a dry run, modify the course ID.
	// This avoids source and image conflicts.
	if options.DryRun {
		course.ID = DRY_RUN_PREFIX + course.ID
	}

	return course, originalID, nil
}
