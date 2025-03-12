package util

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/edulinq/autograder/internal/log"
)

const DEFAULT_MKDIR_PERMS os.FileMode = 0755

var tempDir string = filepath.Join("/", "tmp", "autograder-temp")
var tempDirMutex sync.Mutex
var createdTempDirs []string
var shouldRemoveTempDirs bool = true

func SetTempDirForTesting(newTempDir string) {
	tempDirMutex.Lock()
	defer tempDirMutex.Unlock()

	tempDir = newTempDir
}

func MustMkDirTemp(prefix string) string {
	path, err := MkDirTemp(prefix)
	if err != nil {
		log.Fatal("Failed to create temp path.", log.NewAttr("path", path))
	}

	return path
}

func MkDirTemp(prefix string) (string, error) {
	return MkDirTempFull(prefix, true)
}

func MkDirTempFull(prefix string, cleanupTempDir bool) (string, error) {
	tempDirMutex.Lock()
	defer tempDirMutex.Unlock()

	if tempDir != "" {
		MkDir(tempDir)
	}

	dir, err := os.MkdirTemp(tempDir, prefix)
	if err != nil {
		return "", err
	}

	if cleanupTempDir {
		createdTempDirs = append(createdTempDirs, dir)
	}

	return dir, nil
}

func MkDir(path string) error {
	return MkDirPerms(path, DEFAULT_MKDIR_PERMS)
}

func MustMkDir(path string) {
	err := MkDir(path)
	if err != nil {
		log.Fatal("Failed to create dir.", log.NewAttr("path", path))
	}
}

func MkDirPerms(path string, perms os.FileMode) error {
	return os.MkdirAll(path, perms)
}

func clearRecordedTempDirs() {
	createdTempDirs = nil
}

// Remove all the temp dirs created via MkDirTemp().
func RemoveRecordedTempDirs() error {
	tempDirMutex.Lock()
	defer tempDirMutex.Unlock()

	if !shouldRemoveTempDirs {
		return nil
	}

	var errs error = nil
	for _, dir := range createdTempDirs {
		errs = errors.Join(errs, RemoveDirent(dir))
	}

	clearRecordedTempDirs()

	return errs
}

func SetShouldRemoveTempDirs(shouldRemove bool) {
	tempDirMutex.Lock()
	defer tempDirMutex.Unlock()

	shouldRemoveTempDirs = shouldRemove
}

// If this dir has a single dirent, return its path.
// If there is not exactly one dirent (or this is not a dir), then return an empty string.
func GetSingleDirent(dir string) string {
	if !IsDir(dir) {
		return ""
	}

	dirents, err := os.ReadDir(dir)
	if err != nil {
		log.Warn("Could not list dir for single dirent.", err, log.NewAttr("dir", dir))
		return ""
	}

	if len(dirents) != 1 {
		return ""
	}

	return filepath.Join(dir, dirents[0].Name())
}

// Take in two dirs and list all the files that match by relative path and don't match.
// Matches will be specified by just listing the matching relative paths.
// Non-matches will be specified with a slice of arrays, where the index of the dir with the file will have the relative path and the other will have an empty string.
func MatchFiles(dirs [2]string) ([]string, [][2]string, error) {
	// Track which dirs have each file.
	allFiles := map[string][]bool{}

	for i, dir := range dirs {
		relpaths, err := GetAllDirents(dirs[i], true, true)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to get relative files for '%s': '%w'.", dir, err)
		}

		for _, relpath := range relpaths {
			_, exists := allFiles[relpath]
			if !exists {
				allFiles[relpath] = make([]bool, len(dirs))
			}

			allFiles[relpath][i] = true
		}
	}

	// Sort the keys.
	sortedKeys := make([]string, 0, len(allFiles))
	for relpath, _ := range allFiles {
		sortedKeys = append(sortedKeys, relpath)
	}
	slices.Sort(sortedKeys)

	// Collect the results.
	matches := []string{}
	unmatches := [][len(dirs)]string{}

	for _, relpath := range sortedKeys {
		matchInfo := allFiles[relpath]

		match := true
		var unmatchRow [len(dirs)]string

		for i, value := range matchInfo {
			match = match && value
			if value {
				unmatchRow[i] = relpath
			}
		}

		if match {
			matches = append(matches, relpath)
		} else {
			unmatches = append(unmatches, unmatchRow)
		}
	}

	return matches, unmatches, nil
}

// Recursively get all the dirents starting with some path (not including that path).
// If the base path is a file, and empty slice will be returned.
// Set |relPaths| to return relative instead of absolute paths.
// Set |onlyFiles| to only return dirents that are not dirs.
func GetAllDirents(basePath string, relPaths bool, onlyFiles bool) ([]string, error) {
	basePath = ShouldAbs(basePath)
	paths := make([]string, 0)

	if IsFile(basePath) {
		return paths, nil
	}

	err := filepath.WalkDir(basePath, func(path string, dirent fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if basePath == path {
			return nil
		}

		if onlyFiles && dirent.IsDir() {
			return nil
		}

		if relPaths {
			path = RelPath(path, basePath)
		}

		paths = append(paths, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	slices.Sort(paths)

	return paths, nil
}
