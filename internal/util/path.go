package util

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/edulinq/autograder/internal/log"
)

// Get an absolute path of a path.
// On error, log the error and return the original path.
// filepath.Abs() errors out when the path is not abs and the cwd cannot be fetched
// (like if our cwd has been deleted from under us).
// We will just treat this as a fatal error.
func ShouldAbs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Error("Failed to compute an absolute path.", err, log.NewAttr("path", path))
		return path
	}

	return absPath
}

// Tell if a path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func IsFile(path string) bool {
	if !PathExists(path) {
		return false
	}

	return !IsDir(path)
}

func IsDir(path string) bool {
	if !PathExists(path) {
		return false
	}

	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		log.Warn("Unexpected error when checking if a path is a dir.", err, log.NewAttr("path", path))
		return false
	}

	return stat.IsDir()
}

func IsEmptyDir(path string) bool {
	if !IsDir(path) {
		return false
	}

	dir, err := os.Open(path)
	if err != nil {
		log.Warn("Failed to open dir to check if it is empty.", err, log.NewAttr("path", path))
		return false
	}
	defer dir.Close()

	_, err = dir.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			return true
		}

		log.Warn("Unexpected error when reading dir names.", err, log.NewAttr("path", path))
		return false
	}

	return false
}

func IsSymLink(path string) bool {
	if !PathExists(path) {
		return false
	}

	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		log.Warn("Unexpected error when checking if a path is a symbolic link.", err, log.NewAttr("path", path))
		return false
	}

	return (stat.Mode() & fs.ModeSymlink) != 0
}

func FindFiles(filename string, dir string) ([]string, error) {
	return FindDirents(filename, dir, true, false, true)
}

// When symbolic links are allowed, keep two things in mind:
//  1. A retutned path may be outside the passed in dir.
//  2. This method will not terminate if there are loops.
//
// If the filename is empty, return all dirents.
func FindDirents(filename string, dir string, allowFiles bool, allowDirs bool, allowLinks bool) ([]string, error) {
	matches := make([]string, 0)

	if !IsDir(dir) {
		return matches, nil
	}

	err := filepath.WalkDir(dir, func(path string, dirent fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dirent.Type()&fs.ModeSymlink != 0 {
			// Dirent is a link.
			if !allowLinks {
				return nil
			}

			path, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}

			newEntries, err := FindDirents(filename, path, allowFiles, allowDirs, allowLinks)
			if err != nil {
				return err
			}

			matches = append(matches, newEntries...)
			return nil
		}

		if dirent.IsDir() {
			if !allowDirs {
				return nil
			}
		} else {
			if !allowFiles {
				return nil
			}
		}

		if (filename == "") || (filename == dirent.Name()) {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Encountered error(s) while walking course dirs ('%s'): '%w'.", dir, err)
	}

	return matches, nil
}

// Get the directory of the source file calling this method.
func ShouldGetThisDir() string {
	// 0 is the current caller (this function), and 1 should be one frame back.
	_, path, _, ok := runtime.Caller(1)
	if !ok {
		log.Error("Could not get the stack frame for the current runtime.")
		return "."
	}

	return filepath.Dir(path)
}

// Check this directory and all parent directories for a file with a specific name.
// If nothing is found, an empty string will be returned.
func SearchParents(basepath string, name string) string {
	basepath = ShouldAbs(basepath)

	if IsFile(basepath) {
		basepath = filepath.Dir(basepath)
	}

	for {
		targetPath := filepath.Join(basepath, name)

		if !PathExists(targetPath) {
			// Move up to the parent.
			oldBasepath := basepath
			basepath = filepath.Dir(basepath)

			if oldBasepath == basepath {
				break
			}

			continue
		}

		return targetPath
	}

	return ""
}

// Return true if the path has the given parent (or is the parent itself).
// If the suspected parent dir does not actually exist, false will be returned.
// The suspected child path can have terminal components that do not exist,
// but a parent must eventually match the suspected parent for true to be returned.
// This method will recursively walk up the path and its parents to see if they match the suspected parent.
func PathHasParentOrSelf(path string, parent string) bool {
	if !PathExists(parent) {
		return false
	}

	path = ShouldNormalizePath(path)
	parent = ShouldNormalizePath(parent)

	if ShouldSameDirent(path, parent) {
		return true
	}

	// Check the path's parent.
	newPath := filepath.Dir(path)

	// If the path and new path match, then we have hit (and already checked) the root.
	if ShouldSameDirent(path, newPath) {
		return false
	}

	return PathHasParentOrSelf(newPath, parent)
}

// Get the child's path relative to the parent.
// This operation is purely lexical and can fail in non-trivial situations (e.g., hard links or strange mounts).
// If the paths are the same, then "." will be returned.
// If the path is not a child of the path, then a cleaned absolute version of the path will be returned.
// This method is not robust (in several ways) and should not be used with user-supplied paths.
func RelPath(child string, parent string) string {
	child = filepath.Clean(ShouldAbs(child))
	parent = filepath.Clean(ShouldAbs(parent)) + string(os.PathSeparator)

	if child == parent {
		return "."
	}

	return strings.TrimPrefix(child, parent)
}

// Get the root directory of this project.
// This is decently fragile and can easily break in a deployment/production setting.
// Should only be used for testing purposes.
func RootDirForTesting() string {
	return ShouldAbs(filepath.Join(ShouldAbs(ShouldGetThisDir()), "..", ".."))
}

func TestdataDirForTesting() string {
	return filepath.Join(RootDirForTesting(), "testdata")
}

// Return path if it is absolute,
// otherwise return the join of baseDir and path.
func JoinIfNotAbs(path string, baseDir string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(baseDir, path)
}

// Normalize a path as best as possible.
func ShouldNormalizePath(path string) string {
	path = filepath.Clean(path)
	path = ShouldAbs(path)

	if !PathExists(path) {
		return path
	}

	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		log.Debug("Failed to eval symlinks for realpath.", err, log.NewAttr("path", path))
		return path
	}

	return realPath
}

// Best effort attempt to determine the absolute path to the directory that holds a custom package path.
// The package path is expected to be an internal path of the form `github.com/edulinq/autograder/`.
func GetDirPathFromCustomPackagePath(packagePath string) string {
	if strings.HasPrefix(packagePath, "github.com/edulinq/autograder/") {
		packagePath = strings.TrimPrefix(packagePath, "github.com/edulinq/autograder/")
	}

	parts := strings.Split(packagePath, "/")

	// Package paths are relative from the base directory `autograder-server`.
	path := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", ".."))
	for _, part := range parts {
		path = filepath.Join(path, part)
	}

	return path
}
