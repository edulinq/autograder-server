package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/docker/docker/api/types/image"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

type ImageSource interface {
	log.Loggable

	FullID() string
	GetSourceDir() string
	GetCachePath() string
	GetFileCachePath() string
	GetImageInfo() *ImageInfo
	GetImageLock() *sync.Mutex
}

const CACHE_KEY_BUILD_SUCCESS = "image-build-success"

func BuildImageFromSourceQuick(imageSource ImageSource) error {
	return BuildImageFromSource(imageSource, false, true, NewBuildOptions())
}

func BuildImageFromSource(imageSource ImageSource, force bool, quick bool, options *BuildOptions) error {
	if config.DOCKER_DISABLE.Get() {
		return nil
	}

	imageSource.GetImageLock().Lock()
	defer imageSource.GetImageLock().Unlock()

	if force {
		quick = false
	}

	build, err := NeedImageRebuild(imageSource, quick)
	if err != nil {
		return fmt.Errorf("Could not check if image needs building for image source '%s': '%w'.", imageSource.FullID(), err)
	}

	if !force && !build {
		// Nothing has changed, skip build.
		log.Debug("No files have changed, skipping image build.", imageSource)
		return nil
	}

	buildErr := BuildImageWithOptions(imageSource, options)

	// Always try to store the result of cache building.
	_, _, cacheErr := util.CachePut(imageSource.GetCachePath(), CACHE_KEY_BUILD_SUCCESS, (buildErr == nil))

	return errors.Join(buildErr, cacheErr)
}

func NeedImageRebuild(imageSource ImageSource, quick bool) (bool, error) {
	// Check if the last build failed.
	lastBuildSuccess, exists, err := util.CacheFetch(imageSource.GetCachePath(), CACHE_KEY_BUILD_SUCCESS)
	if err != nil {
		return false, fmt.Errorf("Failed to fetch the last build status from cahce for image source '%s': '%w'.", imageSource.FullID(), err)
	}

	lastBuildFailed := true
	if exists {
		lastBuildFailed = !(lastBuildSuccess.(bool))
	}

	if lastBuildFailed && quick {
		return true, nil
	}

	// Check if the image info has changed.
	imageInfoHash, err := util.MD5StringHex(util.MustToJSON(imageSource.GetImageInfo()))
	if err != nil {
		return false, fmt.Errorf("Failed to hash image info for image source '%s': '%w'.", imageSource.FullID(), err)
	}

	oldHash, _, err := util.CachePut(imageSource.GetCachePath(), "image-info", imageInfoHash)
	if err != nil {
		return false, fmt.Errorf("Failed to put image info hash into cahce for image source '%s': '%w'.", imageSource.FullID(), err)
	}

	imageInfoHashHasChanges := (imageInfoHash != oldHash)
	if imageInfoHashHasChanges && quick {
		return true, nil
	}

	// Check if the static files have changes.
	staticFilesHaveChanges, err := CheckFileChanges(imageSource, quick)
	if err != nil {
		return false, fmt.Errorf("Could not check if static files changed for image source '%s': '%w'.", imageSource.FullID(), err)
	}

	return (lastBuildFailed || imageInfoHashHasChanges || staticFilesHaveChanges), nil
}

// Check if the imageSource's static files have changes since the last time they were cached.
// This is thread-safe.
func CheckFileChanges(imageSource ImageSource, quick bool) (bool, error) {
	baseDir := imageSource.GetSourceDir()

	fileCachePath := imageSource.GetFileCachePath()
	cachePath := imageSource.GetCachePath()

	imageInfo := imageSource.GetImageInfo()

	paths := make([]string, 0, len(imageInfo.StaticFiles))
	gitChanges := false

	for _, filespec := range imageInfo.StaticFiles {
		if quick && gitChanges {
			return true, nil
		}

		switch filespec.Type {
		case util.FILESPEC_TYPE_EMPTY, util.FILESPEC_TYPE_NIL, util.FILESPEC_TYPE_URL:
			// no-op.
			continue
		case util.FILESPEC_TYPE_PATH:
			// Collect paths to test all at once.
			paths = append(paths, filepath.Join(baseDir, filespec.GetPath()))
		case util.FILESPEC_TYPE_GIT:
			// Check git refs for changes.

			if filespec.Reference == "" {
				log.Warn("Git repo without ref (branch/commit) used as a static file. Please specify a ref so changes can be seen.",
					imageSource, log.NewAttr("repo", filespec.GetPath()))
			}

			oldRef, exists, err := util.CachePut(cachePath, filespec.GetPath(), filespec.Reference)
			if err != nil {
				return false, err
			}

			if !exists || (oldRef != filespec.Reference) {
				gitChanges = true
			}
		default:
			return false, fmt.Errorf("Unknown filespec type '%s': '%v'.", filespec, filespec.Type)
		}
	}

	pathChanges, err := util.CheckFileChanges(fileCachePath, paths, quick)
	if err != nil {
		return false, err
	}

	return (gitChanges || pathChanges), nil
}

// Make sure the image (which should include the verion) exists.
// If it does not exist, it will be pulled.
// To check if the image is listed, the RepoTags field will be checked for the image's name.
func EnsureImage(name string) error {
	docker, err := getDockerClient()
	if err != nil {
		return err
	}
	defer docker.Close()

	images, err := docker.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list docker images: '%w'.", err)
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == name {
				return nil
			}
		}
	}

	log.Debug("Did not find image locally, attempting pull.", log.NewAttr("name", name))

	return PullImage(name)
}

func PullImage(name string) error {
	docker, err := getDockerClient()
	if err != nil {
		return err
	}
	defer docker.Close()

	output, err := docker.ImagePull(context.Background(), name, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("Failed to run docker image pull command: '%w'.", err)
	}
	defer output.Close()

	var buffer bytes.Buffer
	io.Copy(&buffer, output)

	log.Debug("Image Pull", log.NewAttr("name", name), log.NewAttr("output", buffer.String()))

	return nil
}
