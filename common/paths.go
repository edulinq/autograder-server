package common

import (
    "os"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

func ShouldGetCWD() string {
    if (config.TESTING_MODE.Get()) {
        return util.RootDirForTesting();
    }

    cwd, err := os.Getwd();
    if (err != nil) {
        log.Error().Err(err).Msg("Failed to get working directory.");
        return ".";
    }

    return cwd;
}
