package util

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/log"
)

const (
	UNKNOWN_VERSION  string = "???"
	UNKNOWN_HASH     string = "????????"
	VERSION_FILENAME string = "VERSION.json"
	DIRTY_SUFFIX     string = "dirty"
	UNKNOWN_API      int    = -1
	HASH_LENGTH      int    = 8
)

type Version struct {
	Short   string `json:"short-version"`
	Hash    string `json:"git-hash"`
	IsDirty bool   `json:"is-dirty"`
	Api     int    `json:"api-version"`
}

func (this Version) String() string {
	fullVersion := this.Short + "-" + this.Hash
	if this.IsDirty {
		fullVersion = fullVersion + "-" + DIRTY_SUFFIX
	}

	return fullVersion
}

func MustGetAPIVersion() int {
	version, err := readVersion()
	if err != nil {
		log.Fatal("Failed to read the version from JSON file.", err)
	}

	return version.Api
}

func readVersion() (Version, error) {
	version := Version{
		Short:   UNKNOWN_VERSION,
		Hash:    UNKNOWN_HASH,
		IsDirty: true,
		Api:     UNKNOWN_API,
	}

	versionPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", "..", "resources", VERSION_FILENAME))
	if !IsFile(versionPath) {
		return version, fmt.Errorf("Version path '%s' does not exist.", versionPath)
	}

	err := JSONFromFile(versionPath, &version)
	if err != nil {
		return version, err
	}

	return version, nil
}

func getGitInfo() (string, bool, error) {
	repoPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", ".."))

	hash, err := GitGetCommitHash(repoPath)
	if err != nil {
		return UNKNOWN_HASH, true, err
	}

	isDirty, err := GitRepoIsDirtyHack(repoPath)
	if err != nil {
		return hash[0:HASH_LENGTH], true, err
	}

	return hash[0:HASH_LENGTH], isDirty, nil
}

func GetAutograderVersion() (Version, error) {
	var gitErr error
	version, versionErr := readVersion()

	if version.Hash == UNKNOWN_HASH {
		version.Hash, version.IsDirty, gitErr = getGitInfo()
	}

	return version, errors.Join(versionErr, gitErr)
}
