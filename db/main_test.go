package db

import (
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    // Run inside a func so defers will run before os.Exit().
    code := func() int {
        PrepForTestingMain();
        defer CleanupTestingMain();

        return suite.Run();
    }();

    os.Exit(code);
}
