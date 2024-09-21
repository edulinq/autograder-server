package util

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/log"
)

const (
	UNKNOWN_VERSION  string = "???"
	UNKNOWN_HASH     string = "????????"
	VERSION_FILENAME string = "VERSION.json"
	DIRTY_SUFFIX     string = "dirty"
	UNKNOWN_API      int    = 0
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
	version, err := versionFromJSON()
	if err != nil {
		log.Fatal("Failed to read the version from JSON file.", err)
	}

	return version.Api
}

func versionFromJSON() (Version, error) {
	var version Version

	versionPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", "..", VERSION_FILENAME))
	if !IsFile(versionPath) {
		return version, fmt.Errorf("Version path '%s' does not exist.", versionPath)
	}

	err := JSONFromFile(versionPath, &version)
	if err != nil {
		return version, fmt.Errorf("Failed to read the version from JSON file '%s'.", versionPath)
	}

	return version, nil
}

func getGitInfo() (string, bool) {
	repoPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", ".."))

	hash, err := GitGetCommitHash(repoPath)
	if err != nil {
		log.Error("Failed to get commit hash.", err, log.NewAttr("path", repoPath))
		hash = UNKNOWN_HASH
	}

	isDirty, err := GitRepoIsDirtyHack(repoPath)
	if err != nil {
		log.Error("Failed to get state of the git repository ", err, log.NewAttr("path", repoPath))
		return hash[0:HASH_LENGTH], true
	}

	return hash[0:HASH_LENGTH], isDirty
}

func GetAutograderVersion() (Version, error) {
	version, err := versionFromJSON()
	if err != nil {
		return version, err
	}

	if version.Hash == "" {
		gitHash, isDirty := getGitInfo()

		version.Hash = gitHash
		version.IsDirty = isDirty
	}

	return version, nil
}
