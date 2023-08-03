package util

import (
    "fmt"
    "io/fs"
    "path/filepath"
    "os"

    "github.com/rs/zerolog/log"
)

// filepath.Abs() errors out when the path is not abs and the cwd cannot be fetched
// (like if our cwd has been deleted from under us).
// We will just treat this as a fatal error.
func MustAbs(path string) string {
    absPath, err := filepath.Abs(path);
    if (err != nil) {
        log.Fatal().Str("path", path).Err(err).Msg("Failed to compute an absolute path.");
    }

    return absPath;
}

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

func FindFiles(filename string, dir string) ([]string, error) {
    matches := make([]string, 0);

    if (!IsDir(dir)) {
        return matches, nil;
    }

    err := filepath.WalkDir(dir, func(path string, dirent fs.DirEntry, err error) error {
        if (err != nil) {
            return err;
        }

        if (filename == dirent.Name()) {
            matches = append(matches, path);
        }

        return nil;
    });

    if (err != nil) {
        return nil, fmt.Errorf("Encountered error(s) while walking course dirs ('%s'): '%w'.", dir, err);
    }

    return matches, nil;
}
