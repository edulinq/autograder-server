package util

import (
    "fmt"
    "io"
    "io/fs"
    "path/filepath"
    "os"
    "runtime"
    "strings"

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

func IsFile(path string) bool {
    if (!PathExists(path)) {
        return false;
    }

    return !IsDir(path);
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

        log.Warn().Err(err).Str("path", path).Msg("Unexpected error when checking if a path is a dir.");
        return false;
    }

    return stat.IsDir();
}

func IsEmptyDir(path string) bool {
    if (!IsDir(path)) {
        return false;
    }

    dir, err := os.Open(path);
    if (err != nil) {
        log.Warn().Err(err).Str("path", path).Msg("Failed to open dir to check if it is empty.");
        return false;
    }
    defer dir.Close();

    _, err = dir.Readdirnames(1);
    if (err != nil) {
        if (err == io.EOF) {
            return true;
        }

        log.Warn().Err(err).Str("path", path).Msg("Unexpected error when reading dir names.");
        return false;
    }

    return false;
}

func FindFiles(filename string, dir string) ([]string, error) {
    return FindDirents(filename, dir, true, false, true);
}

// When symbolic links are allowed, keep two things in mind:
//  1) A retutned path may be outside the passed in dir.
//  2) This method will not terminate if there are loops.
func FindDirents(filename string, dir string, allowFiles bool, allowDirs bool, allowLinks bool) ([]string, error) {
    matches := make([]string, 0);

    if (!IsDir(dir)) {
        return matches, nil;
    }

    err := filepath.WalkDir(dir, func(path string, dirent fs.DirEntry, err error) error {
        if (err != nil) {
            return err;
        }

        if (dirent.Type() & fs.ModeSymlink != 0) {
            // Dirent is a link.
            if (!allowLinks) {
                return nil;
            }

            path, err = filepath.EvalSymlinks(path);
            if (err != nil) {
                return err;
            }

            newEntries, err := FindDirents(filename, path, allowFiles, allowDirs, allowLinks);
            if (err != nil) {
                return err;
            }

            matches = append(matches, newEntries...);
            return nil;
        }

        if (dirent.IsDir()) {
            if (!allowDirs) {
                return nil;
            }
        } else {
            if (!allowFiles) {
                return nil;
            }
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

// Get the directory of the source file calling this method.
func GetThisDir() string {
    // 0 is the current caller (this function), and 1 should be one frame back.
    _, path, _, ok := runtime.Caller(1);
    if (!ok) {
        log.Fatal().Msg("Could not get the stackframe for the current runtime.");
    }

    return filepath.Dir(path);
}

// Check this directory and all parent directories for a file with a specific name.
// If nothing is found, an empty string will be returned.
func SearchParents(basepath string, name string) string {
    basepath = MustAbs(basepath);

    for ; ; {
        targetPath := filepath.Join(basepath, name);

        if (!PathExists(targetPath)) {
            // Move up to the parent.
            oldBasepath := basepath;
            basepath = filepath.Dir(basepath);

            if (oldBasepath == basepath) {
                break;
            }

            continue;
        }

        return targetPath;
    }

    return "";
}

// This method is not robust (in many ways) and should be generally avoided in non-testing code.
func PathHasParent(child string, parent string) bool {
    child = MustAbs(child);
    parent = MustAbs(parent);

    return strings.HasPrefix(child, parent);
}

// Get the root directory of this project.
// This is decently fragile and can easily break in a deployment/production setting.
// Should only be used for testing purposes.
func RootDirForTesting() string {
    return MustAbs(filepath.Join(MustAbs(GetThisDir()), ".."));
}
