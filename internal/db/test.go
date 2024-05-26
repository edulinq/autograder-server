package db

import (
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
	MustAddCourses()
	MustAddTestUsers()
}

func CleanupTestingMain() {
	MustClose()

	// Remove any temp directories.
	err := util.RemoveRecordedTempDirs()
	if err != nil {
		log.Error("Error when removing temp dirs.", err)
	}
}

func MustAddTestUsers() {
	path := filepath.Join(config.GetCourseImportDir(), "testdata", model.USERS_FILENAME)

	users, err := model.LoadServerUsersFile(path)
	if err != nil {
		log.Fatal("Could not open test users file.", err, log.NewAttr("path", path))
	}

	err = UpsertUsers(users)
	if err != nil {
		log.Fatal("Failed to upsert test users.", err)
	}
}
