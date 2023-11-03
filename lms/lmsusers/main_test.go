package lmsusers

import (
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    config.EnableTestingMode(false, true);

    os.Exit(suite.Run())
}
