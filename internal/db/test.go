package db

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const TEST_COURSE_ID = "course101"
const TEST_ASSIGNMENT_ID = "hw0"

func MustGetTestCourse() *model.Course {
	return MustGetCourse(TEST_COURSE_ID)
}

func MustGetTestAssignment() *model.Assignment {
	return MustGetAssignment(TEST_COURSE_ID, TEST_ASSIGNMENT_ID)
}

// Perform the standard actions that prep for a package's testing main.
// Callers should make sure to cleanup after testing:
// `defer db.CleanupTestingMain();`.
func PrepForTestingMain() {
	config.MustEnableUnitTestingMode()

	// Quiet the logs.
	log.SetLevelFatal()

	MustOpen()

	ResetForTesting()
}

// A reset function than can be called between tests.
func ResetForTesting() {
	MustClear()
	MustClose()

	// Open will add test data when in testing mode.
	MustOpen()
}

func CleanupTestingMain() {
	MustClose()

	// Remove any temp directories.
	err := util.RemoveRecordedTempDirs()
	if err != nil {
		log.Error("Error when removing temp dirs.", err)
	}
}

func addTestCourse(configPath string) (*model.Course, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	// First add the test course.
	course, err := backend.AddTestCourse(configPath)
	if err != nil {
		return nil, err
	}

	// Second, copy over the source.
	// Note that the timing here may seem off,
	// but we need the course first to know what the source dir will be.

	// Copy over the source directory.
	baseDir := filepath.Dir(configPath)
	sourceDir := course.GetBaseSourceDir()

	if util.PathExists(sourceDir) {
		err := util.RemoveDirent(sourceDir)
		if err != nil {
			return nil, fmt.Errorf("Failed to remove source dir '%s': '%w'.", sourceDir, err)
		}
	}

	err = util.CopyDirWhole(baseDir, sourceDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy source dir from '%s' to '%s': '%w'.", baseDir, sourceDir, err)
	}

	return course, nil
}

// Add all the test courses to the database.
// As a test function, this will take several shortcuts.
// Courses should only officially be added via the procedures/courses package.
func addTestCourses() error {
	baseDir := config.GetTestdataDir()

	configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir)
	if err != nil {
		return fmt.Errorf("Failed to search for test course configs in '%s': '%w'.", baseDir, err)
	}

	for _, configPath := range configPaths {
		_, err := addTestCourse(configPath)
		if err != nil {
			return fmt.Errorf("Failed to add course from config '%s': '%w'.", configPath, err)
		}
	}

	return nil
}

func addTestUsers() error {
	path := filepath.Join(config.GetTestdataDir(), model.USERS_FILENAME)

	users, err := model.LoadServerUsersFile(path)
	if err != nil {
		return fmt.Errorf("Could not open test users file '%s': '%w'.", path, err)
	}

	err = UpsertUsers(users)
	if err != nil {
		return fmt.Errorf("Failed to upsert test users: '%w'.", err)
	}

	return nil
}
