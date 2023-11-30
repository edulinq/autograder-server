package db

import (
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const TEST_COURSE_ID = "COURSE101";

func MustGetTestCourse() *model.Course {
    return MustGetCourse(TEST_COURSE_ID);
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
    MustLoadCourses();
}

func CleanupTestingMain() {
    MustClose();

    // Remove the temp working directory set in config.MustEnableUnitTestingMode().
    util.RemoveDirent(config.WORK_DIR.Get());
}
