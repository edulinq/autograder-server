package util

import (
    "path/filepath"
    "strings"

    "github.com/eriq-augustine/autograder/log"
)

const (
    UNKNOWN_VERSION string = "???"
    VERSION_FILENAME string = "VERSION.txt"
    DIRTY_SUFFIX string = "dirty"
    HASH_LENGTH int = 8
)

func GetAutograderVersion() string {
    versionPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", VERSION_FILENAME));
    if (!IsFile(versionPath)) {
        log.Error("Version file does not exist.", log.NewAttr("path", versionPath));
        return UNKNOWN_VERSION;
    }

    version, err := ReadFile(versionPath);
    if (err != nil) {
        log.Error("Failed to read the version file.", err, log.NewAttr("path", versionPath));
        return UNKNOWN_VERSION;
    }

    return strings.TrimSpace(version);
}

func GetAutograderFullVersion() string {
    repoPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), ".."));

    version := GetAutograderVersion();

    hash, err := GitGetCommitHash(repoPath);
    if (err != nil) {
        log.Error("Failed to get commit hash.", err, log.NewAttr("path", repoPath));
        hash = UNKNOWN_VERSION;
    }

    dirtySuffix := "";

    isDirty, err := GitRepoIsDirtyHack(repoPath);
    if (err != nil) {
        dirtySuffix = "-" + UNKNOWN_VERSION;
    }

    if (isDirty) {
        dirtySuffix = "-" + DIRTY_SUFFIX;
    }

    return version + "-" + hash[0:HASH_LENGTH] + dirtySuffix;
}
