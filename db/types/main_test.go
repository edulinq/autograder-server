package types

import (
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

// Note that this is a duplicate of db/test.go, but we have to avoid an import cycle.
const TEST_COURSE_ID = "COURSE101";

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    config.MustEnableUnitTestingMode();

    // Remove the temp working directory set in config.MustEnableUnitTestingMode().
    defer util.RemoveDirent(config.WorkDir.Get());

    os.Exit(suite.Run())
}
