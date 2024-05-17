package util

import (
    "errors"
    "os"
    "sync"
)

const DEFAULT_MKDIR_PERMS os.FileMode = 0755;

var tempDir string = "";
var tempDirMutex sync.Mutex;
var createdTempDirs []string;

func SetTempDirForTesting(newTempDir string) {
    tempDir = newTempDir;
}

func MkDirTemp(prefix string) (string, error) {
    tempDirMutex.Lock();
    defer tempDirMutex.Unlock();

    dir, err := os.MkdirTemp(tempDir, prefix);
    if (err != nil) {
        return "", err;
    }

    createdTempDirs = append(createdTempDirs, dir);
    return dir, nil;
}

func MkDir(path string) error {
    return MkDirPerms(path, DEFAULT_MKDIR_PERMS);
}

func MkDirPerms(path string, perms os.FileMode) error {
    return os.MkdirAll(path, perms);
}

func ClearRecordedTempDirs() {
    createdTempDirs = nil;
}

// Remove all the temp dirs created via MkDirTemp().
func RemoveRecordedTempDirs() error {
    tempDirMutex.Lock();
    defer tempDirMutex.Unlock();

    var errs error = nil;
    for _, dir := range createdTempDirs {
        errs = errors.Join(errs, RemoveDirent(dir));
    }

    ClearRecordedTempDirs();

    return errs;
}
