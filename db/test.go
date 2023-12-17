package db

import (
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const TEST_COURSE_ID = "COURSE101";
const TEST_ASSIGNMENT_ID = "hw0";

func MustGetTestCourse() *model.Course {
    return MustGetCourse(TEST_COURSE_ID);
}

func MustGetTestAssignment() *model.Assignment {
    return MustGetAssignment(TEST_COURSE_ID, TEST_ASSIGNMENT_ID);
}

// Perform the standard actions that prep for a package's testing main.
// Callers should make sure to cleanup after testing:
// `defer db.CleanupTestingMain();`.
func PrepForTestingMain() {
    config.MustEnableUnitTestingMode();

    // Quiet the logs.
    config.SetLogLevelFatal();

    MustOpen();

    ResetForTesting();
}

// A reset function than can be called between tests.
func ResetForTesting() {
    MustClear();
    MustAddCourses();
}

func CleanupTestingMain() {
    MustClose();

    // Remove any temp directories.
    err := util.RemoveRecordedTempDirs();
    if (err != nil) {
        log.Error().Err(err).Msg("Error when removing temp dirs.");
    }
}
