package util

import (
    "os"
)

const DEFAULT_MKDIR_PERMS os.FileMode = 0755;

func MkDir(path string) error {
    return MkDirPerms(path, DEFAULT_MKDIR_PERMS);
}

func MkDirPerms(path string, perms os.FileMode) error {
    return os.MkdirAll(path, perms);
}
