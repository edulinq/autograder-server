package tasks

// Note that this file is largely a copy of db/test.go.
// The content is repeated to avoid an import cycle.

import (
    "os"
    "testing"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    // Run inside a func so defers will run before os.Exit().
    code := func() int {
        config.MustEnableUnitTestingMode();

        defer CleanupTestingMain();

        // Quiet the logs.
        config.SetLogLevelFatal();

        return suite.Run();
    }();

    os.Exit(code);
}

func CleanupTestingMain() {
    // Remove any temp directories.
    err := util.RemoveRecordedTempDirs();
    if (err != nil) {
        log.Error().Err(err).Msg("Error when removing temp dirs.");
    }
}
