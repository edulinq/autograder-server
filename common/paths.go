package common

import (
    "os"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/util"
)

func ShouldGetCWD() string {
    if (config.TESTING_MODE.Get()) {
        return util.RootDirForTesting();
    }

    cwd, err := os.Getwd();
    if (err != nil) {
        log.Error("Failed to get working directory.", err);
        return ".";
    }

    return cwd;
}
