package util

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/edulinq/autograder/internal/log"
)

const (
	UNKNOWN_HASH      string = "????????"
	VERSION_FILENAME  string = "VERSION.json"
	DIRTY_SUFFIX      string = "dirty"
	UNKNOWN_COMPONENT int    = -1
	HASH_LENGTH       int    = 8

	API_DESCRIPTION_FILENAME = "api.json"
	RESOURCES_DIRNAME        = "resources"
)

var (
	cacheLock     sync.Mutex
	cachedVersion *Version = nil
)

type Version struct {
	Base    string `json:"base-version"`
	Hash    string `json:"git-hash"`
	IsDirty bool   `json:"is-dirty"`
}

func (this Version) String() string {
	fullVersion := this.Base + "-" + this.Hash
	if this.IsDirty {
		fullVersion = fullVersion + "-" + DIRTY_SUFFIX
	}

	return fullVersion
}

func (this Version) Major() (int, error) {
	parts := strings.Split(this.Base, ".")
	if len(parts) == 0 {
		return UNKNOWN_COMPONENT, fmt.Errorf("Could not split major version from base version: '%s'.", this.Base)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return UNKNOWN_COMPONENT, fmt.Errorf("Failed for convert major version to int '%s': '%w'.", parts[0], err)
	}

	return major, nil
}

func readVersion() (Version, error) {
	version := Version{
		Base:    fmt.Sprintf("%d.%d.%d", UNKNOWN_COMPONENT, UNKNOWN_COMPONENT, UNKNOWN_COMPONENT),
		Hash:    UNKNOWN_HASH,
		IsDirty: true,
	}

	versionPath := ShouldAbs(filepath.Join(GetResourcesDir(), VERSION_FILENAME))
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

func GetFullCachedVersion() (Version, error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if cachedVersion != nil {
		return *cachedVersion, nil
	}

	version, err := readVersion()
	if err != nil {
		return Version{}, err
	}

	// If there is no hash already set, try to discover it from git.
	if version.Hash == UNKNOWN_HASH {
		version.Hash, version.IsDirty, err = getGitInfo()
		if err != nil {
			return Version{}, fmt.Errorf("Failed to discover Git version information: '%w'.", err)
		}
	}

	cachedVersion = &version

	return version, nil
}

func MustGetAPIVersion() int {
	version, err := GetFullCachedVersion()
	if err != nil {
		log.Fatal("Failed to get the full cached version.", err)
	}

	major, err := version.Major()
	if err != nil {
		log.Fatal("Failed to get major (API) verson.", err)
	}

	return major
}

func GetResourcesDir() string {
	return ShouldAbs(filepath.Join(ShouldGetThisDir(), "..", "..", RESOURCES_DIRNAME))
}

func GetAPIDescriptionFilepath() (string, error) {
	apiPath := ShouldAbs(filepath.Join(GetResourcesDir(), API_DESCRIPTION_FILENAME))
	if !IsFile(apiPath) {
		return apiPath, fmt.Errorf("API description path '%s' does not exist.", apiPath)
	}

	return apiPath, nil
}
