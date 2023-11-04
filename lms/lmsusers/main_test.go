package lmsusers

import (
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    config.EnableUnitTestingMode();

    os.Exit(suite.Run())
}
