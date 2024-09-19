package util

import (
	"path/filepath"
	"strings"

	"github.com/edulinq/autograder/internal/log"
)

const (
	UNKNOWN_VERSION  string = "???"
	UNKNOWN_HASH     string = "????????"
	VERSION_FILENAME string = "VERSION.json"
	DIRTY_SUFFIX     string = "dirty"
	HASH_LENGTH      int    = 8
)

type Version struct{
	Short string `json:"short-version"`
	Hash string `json:"git-hash"`
	Status string `json:"status"`
}

func GetAutograderVersion() string {
	versionPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", "..", VERSION_FILENAME))
	if !IsFile(versionPath) {
		log.Error("Version file does not exist.", log.NewAttr("path", versionPath))
		return UNKNOWN_VERSION
	}

	var readVersion Version
	readVersion.Short = UNKNOWN_VERSION

	err := JSONFromFile(versionPath,&readVersion)
	if err != nil {
		log.Error("Failed to read the version from JSON file.", err, log.NewAttr("path", versionPath))
		return UNKNOWN_VERSION
	}

	return strings.TrimSpace(readVersion.Short)
}

func ComputeAutograderFullVersion() (version string, gitHash string, gitStatus string) {
	repoPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", ".."))

	shortVersion := GetAutograderVersion()

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
	}else {
		status = "clean"
	}


	return shortVersion, hash[0:HASH_LENGTH], status
}

func GetAutograderFullVersion() string {
	versionPath := ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", "..", VERSION_FILENAME))
	if !IsFile(versionPath) {
		log.Error("Version file does not exist.", log.NewAttr("path", versionPath))
		return UNKNOWN_VERSION
	}

	var shortVersion string 
	var gitHash string
	var status string 

	var readVersion Version

	err := JSONFromFile(versionPath,&readVersion)
	if err != nil {
		log.Error("Failed to read the version JSON file.", err, log.NewAttr("path", versionPath))
		return UNKNOWN_VERSION
	}

	if readVersion.Hash == "" || readVersion.Status == ""{
		shortVersion , gitHash, status = ComputeAutograderFullVersion()
		return shortVersion + "-" + gitHash + "-" + status
	}

	shortVersion = readVersion.Short
	gitHash = readVersion.Hash
	status = readVersion.Status
	
	return shortVersion + "-" + gitHash + "-" + status
}