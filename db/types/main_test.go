package types

import (
    "path/filepath"
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
)

const TEST_COURSE_ID = "COURSE101";

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    config.EnableUnitTestingMode();

    // Note that this is a duplicate of db/test.go, but we have to avoid an import cycle.
    MustLoadCourse(filepath.Join(config.COURSES_ROOT.Get(), TEST_COURSE_ID, COURSE_CONFIG_FILENAME));

    os.Exit(suite.Run())
}
