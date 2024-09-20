package util

import (
	"fmt"
	"path/filepath"
	"strings"

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
	Short  string `json:"short-version"`
	Hash   string `json:"git-hash"`
	Status bool   `json:"isDirty"`
	Api    int    `json:"api-version"`
}

func (v Version) FullVersion() string {
	fullVersion := v.Short + "-" + v.Hash
	if v.Status {
		fullVersion = fullVersion + "-" + DIRTY_SUFFIX
	}

	return fullVersion
}

func MustGetAPIVersion() int {
	version, err := ReadVersionFromJSON()
	if err != nil {
		log.Fatal("Failed to read the version from JSON file.", err)
	}

	return version.Api
}

func GetAutograderVersion() string {
	version, err := ReadVersionFromJSON()
	if err != nil {
		log.Error("Failed to read the version from JSON file.")
		return UNKNOWN_VERSION
	}

	return strings.TrimSpace(version.Short)
}

func ReadVersionFromJSON() (*Version, error) {
	readVersion := Version{
		Short:  UNKNOWN_VERSION,
		Hash:   UNKNOWN_HASH,
		Status: true,
		Api:    UNKNOWN_API,
	}

	versionPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", "..", VERSION_FILENAME))
	if !IsFile(versionPath) {
		log.Error("Version file does not exist.", log.NewAttr("path", versionPath))
		return &readVersion, fmt.Errorf("Version file does not exist.")
	}

	err := JSONFromFile(versionPath, &readVersion)
	if err != nil {
		log.Error("Failed to read the version from JSON file.", err, log.NewAttr("path", versionPath))
		return &readVersion, fmt.Errorf("Failed to read the version from JSON file %s.", log.NewAttr("path", versionPath))
	}

	return &readVersion, nil
}

func ComputeAutograderFullVersion() (gitHash string, gitStatus bool) {
	repoPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", ".."))

	hash, err := GitGetCommitHash(repoPath)
	if err != nil {
		log.Error("Failed to get commit hash.", err, log.NewAttr("path", repoPath))
		hash = UNKNOWN_HASH
	}

	isDirty, err := GitRepoIsDirtyHack(repoPath)
	if err != nil {
		return hash[0:HASH_LENGTH], true
	}

	return hash[0:HASH_LENGTH], isDirty
}

func GetAutograderFullVersion() Version {
	var gitHash string
	var status bool

	versionOut := Version{
		Short:  UNKNOWN_VERSION,
		Hash:   UNKNOWN_HASH,
		Status: true,
		Api:    UNKNOWN_API,
	}

	readVersion, err := ReadVersionFromJSON()
	if err != nil {
		log.Error("Failed to read the version from JSON file.")
		return versionOut
	}

	if readVersion.Hash == UNKNOWN_HASH {
		gitHash, status = ComputeAutograderFullVersion()

		versionOut.Short = GetAutograderVersion()
		versionOut.Hash = gitHash
		versionOut.Status = status
		versionOut.Api = MustGetAPIVersion()

		return versionOut
	}

	versionOut.Short = readVersion.Short
	versionOut.Hash = readVersion.Hash
	versionOut.Status = readVersion.Status
	versionOut.Api = readVersion.Api

	return versionOut
}
