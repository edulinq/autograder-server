package db

import (
    "path/filepath"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db/types"
    "github.com/eriq-augustine/autograder/model"
    // TEST
    // "github.com/eriq-augustine/autograder/util"
)

const TEST_COURSE_ID = "COURSE101";

// Load the test course into the database.
// Testing mode should already be enabled.
func MustLoadTestCourse() {
    MustLoadCourse(filepath.Join(config.COURSES_ROOT.Get(), TEST_COURSE_ID, types.COURSE_CONFIG_FILENAME));
}

func MustGetTestCourse() model.Course {
    MustLoadTestCourse();
    return MustGetCourse(TEST_COURSE_ID);
}

// Perform the standard actions that prep for a package's testing main.
// Callers should make sure to cleanup after testing:
// `defer db.CleanupTestingMain();`.
func PrepForTestingMain() {
    config.MustEnableUnitTestingMode();

    MustOpen();

    MustClear();
    MustLoadTestCourse();

    // Quiet the logs.
    config.SetLogLevelFatal();
}

func CleanupTestingMain() {
    MustClose();

    // Remove the temp working directory set in config.MustEnableUnitTestingMode().
    // TEST
    // util.RemoveDirent(config.WORK_DIR.Get());
}
