package util

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/edulinq/autograder/internal/log"
)

const DEFAULT_MKDIR_PERMS os.FileMode = 0755

var tempDir string = filepath.Join("/", "tmp", "autograder-temp")
var tempDirMutex sync.Mutex
var createdTempDirs []string

func SetTempDirForTesting(newTempDir string) {
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
	tempDirMutex.Lock()
	defer tempDirMutex.Unlock()

	if tempDir != "" {
		MkDir(tempDir)
	}

	dir, err := os.MkdirTemp(tempDir, prefix)
	if err != nil {
		return "", err
	}

	createdTempDirs = append(createdTempDirs, dir)
	return dir, nil
}

func MkDir(path string) error {
	return MkDirPerms(path, DEFAULT_MKDIR_PERMS)
}

func MkDirPerms(path string, perms os.FileMode) error {
	return os.MkdirAll(path, perms)
}

func ClearRecordedTempDirs() {
	createdTempDirs = nil
}

// Remove all the temp dirs created via MkDirTemp().
func RemoveRecordedTempDirs() error {
	tempDirMutex.Lock()
	defer tempDirMutex.Unlock()

	var errs error = nil
	for _, dir := range createdTempDirs {
		errs = errors.Join(errs, RemoveDirent(dir))
	}

	ClearRecordedTempDirs()

	return errs
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
