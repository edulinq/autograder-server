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
	Status string `json:"status"`
	Api    int    `json:"api-version"`
}

func MustGetAPIVersion() int {
	version, err := ReadVersionFromJSON()
	if err != nil {
		log.Fatal("Failed to read the version from JSON file.")
		return UNKNOWN_API
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
	var readVersion Version
	readVersion.Short = UNKNOWN_VERSION
	readVersion.Hash = UNKNOWN_HASH
	readVersion.Status = UNKNOWN_VERSION
	readVersion.Api = UNKNOWN_API

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

func ComputeAutograderFullVersion() (gitHash string, gitStatus string) {
	repoPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", ".."))

	hash, err := GitGetCommitHash(repoPath)
	if err != nil {
		log.Error("Failed to get commit hash.", err, log.NewAttr("path", repoPath))
		hash = UNKNOWN_HASH
	}

	var status string

	isDirty, err := GitRepoIsDirtyHack(repoPath)
	if err != nil {
		status = UNKNOWN_VERSION
	}

	if isDirty {
		status = DIRTY_SUFFIX
	}

	return hash[0:HASH_LENGTH], status
}

func GetAutograderFullVersion() Version {
	var gitHash string
	var status string

	versionOut := Version{
		Short:  UNKNOWN_VERSION,
		Hash:   UNKNOWN_HASH,
		Status: "",
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
