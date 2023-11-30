package util

import (
    "path/filepath"
    "strings"

    "github.com/rs/zerolog/log"
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
        log.Error().Str("path", versionPath).Msg("Version file does not exist.");
        return UNKNOWN_VERSION;
    }

    version, err := ReadFile(versionPath);
    if (err != nil) {
        log.Error().Err(err).Str("path", versionPath).Msg("Failed to read the version file.");
        return UNKNOWN_VERSION;
    }

    return strings.TrimSpace(version);
}

func GetAutograderFullVersion() string {
    repoPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), ".."));

    version := GetAutograderVersion();

    hash, err := GitGetCommitHash(repoPath);
    if (err != nil) {
        log.Error().Err(err).Str("path", repoPath).Msg("Failed to get commit hash.");
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
