package common

import (
    "os"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/util"
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
