package db

import (
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    config.EnableUnitTestingMode();

    MustOpen();
    MustLoadTestCourse();

    // Quiet the logs.
    config.SetLogLevelFatal();

    os.Exit(suite.Run())
}
