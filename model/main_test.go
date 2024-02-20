package model

// Note that this file is largely a copy of db/test.go.
// The content is repeated to avoid an import cycle.

import (
    "os"
    "testing"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    // Run inside a func so defers will run before os.Exit().
    code := func() int {
        config.MustEnableUnitTestingMode();

        defer CleanupTestingMain();

        return suite.Run();
    }();

    os.Exit(code);
}

func CleanupTestingMain() {
    // Remove any temp directories.
    err := util.RemoveRecordedTempDirs();
    if (err != nil) {
        log.Error("Error when removing temp dirs.", err);
    }
}
