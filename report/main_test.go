package report

import (
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/db"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    // Run inside a func so defers will run before os.Exit().
    code := func() int {
        db.PrepForTestingMain();
        defer db.CleanupTestingMain();

        return suite.Run();
    }();

    os.Exit(code);
}
