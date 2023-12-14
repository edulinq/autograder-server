package common

// Note that this file is largely a copy of db/test.go.
// The content is repeated to avoid an import cycle.

import (
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    // Run inside a func so defers will run before os.Exit().
    code := func() int {
        config.MustEnableUnitTestingMode();

        // Remove the temp working directory set in config.MustEnableUnitTestingMode().
        defer util.RemoveDirent(config.WORK_DIR.Get());

        return suite.Run();
    }();

    os.Exit(code);
}
