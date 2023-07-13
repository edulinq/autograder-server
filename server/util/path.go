package util

import (
    "os"
)

// Tell if a path exists.
func PathExists(path string) bool {
    _, err := os.Stat(path);
    if (err != nil) {
        if os.IsNotExist(err) {
            return false;
        }
    }

    return true;
}

func IsDir(path string) bool {
    if (!PathExists(path)) {
        return false;
    }

    stat, err := os.Stat(path);
    if (err != nil) {
        if os.IsNotExist(err) {
            return false;
        }

        return false;
    }

    return stat.IsDir();
}
