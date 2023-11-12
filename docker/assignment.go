package docker

import (
    "errors"
    "fmt"
    "path/filepath"
    "sync"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

type ImageSource interface {
    GetID() string;
    GetSourceDir() string;
    GetCachePath() string;
    GetFileCachePath() string;
    GetImageInfo() *ImageInfo;
    GetImageLock() *sync.Mutex;
}

const CACHE_KEY_BUILD_SUCCESS = "image-build-success"

func BuildImageFromSourceQuick(imageSource ImageSource) error {
    return BuildImageFromSource(imageSource, false, true, NewBuildOptions());
}

func BuildImageFromSource(imageSource ImageSource, force bool, quick bool, options *BuildOptions) error {
    if (config.DOCKER_DISABLE.Get()) {
        return nil;
    }

    imageSource.GetImageLock().Lock();
    defer imageSource.GetImageLock().Unlock();

    if (force) {
        quick = false;
    }

    build, err := NeedImageRebuild(imageSource, quick);
    if (err != nil) {
        return fmt.Errorf("Could not check if image needs building for image source '%s': '%w'.", imageSource.GetID(), err);
    }

    if (!force && !build) {
        // Nothing has changed, skip build.
        log.Debug().Str("imageSource", imageSource.GetID()).Msg("No files have changed, skipping image build.");
        return nil;
    }

    buildErr := BuildImageWithOptions(imageSource.GetImageInfo(), options);

    // Always try to store the result of cache building.
    _, _, cacheErr := util.CachePut(imageSource.GetCachePath(), CACHE_KEY_BUILD_SUCCESS, (buildErr == nil));

    return errors.Join(buildErr, cacheErr);
}

func NeedImageRebuild(imageSource ImageSource, quick bool) (bool, error) {
    // Check if the last build failed.
    lastBuildSuccess, exists, err := util.CacheFetch(imageSource.GetCachePath(), CACHE_KEY_BUILD_SUCCESS);
    if (err != nil) {
        return false, fmt.Errorf("Failed to fetch the last build status from cahce for image source '%s': '%w'.", imageSource.GetID(), err);
    }

    lastBuildFailed := true;
    if (exists) {
        lastBuildFailed = !(lastBuildSuccess.(bool));
    }

    if (lastBuildFailed && quick) {
        return true, nil;
    }

    // Check if the image info has changed.
    imageInfoHash, err := util.MD5StringHex(util.MustToJSON(imageSource.GetImageInfo()));
    if (err != nil) {
        return false, fmt.Errorf("Failed to hash image info for image source '%s': '%w'.", imageSource.GetID(), err);
    }

    oldHash, _, err := util.CachePut(imageSource.GetCachePath(), "image-info", imageInfoHash);
    if (err != nil) {
        return false, fmt.Errorf("Failed to put image info hash into cahce for image source '%s': '%w'.", imageSource.GetID(), err);
    }

    imageInfoHashHasChanges := (imageInfoHash != oldHash);
    if (imageInfoHashHasChanges && quick) {
        return true, nil;
    }

    // Check if the static files have changes.
    staticFilesHaveChanges, err := CheckFileChanges(imageSource, quick);
    if (err != nil) {
        return false, fmt.Errorf("Could not check if static files changed for image source '%s': '%w'.", imageSource.GetID(), err);
    }

    return (lastBuildFailed || imageInfoHashHasChanges || staticFilesHaveChanges), nil;
}

// Check if the imageSource's static files have changes since the last time they were cached.
// This is thread-safe.
func CheckFileChanges(imageSource ImageSource, quick bool) (bool, error) {
    baseDir := imageSource.GetSourceDir();

    fileCachePath := imageSource.GetFileCachePath();
    cachePath := imageSource.GetCachePath();

    imageInfo := imageSource.GetImageInfo();

    paths := make([]string, 0, len(imageInfo.StaticFiles));
    gitChanges := false;

    for _, filespec := range imageInfo.StaticFiles {
        if (quick && gitChanges) {
            return true, nil;
        }

        switch (filespec.GetType()) {
            case common.FILESPEC_TYPE_PATH:
                // Collect paths to test all at once.
                paths = append(paths, filepath.Join(baseDir, filespec.GetPath()));
            case common.FILESPEC_TYPE_GIT:
                // Check git refs for changes.
                url, _, ref, err := filespec.ParseGitParts();
                if (err != nil) {
                    return false, err;
                }

                if (ref == "") {
                    log.Warn().Str("imageSource", imageSource.GetID()).Str("repo", url).
                            Msg("Git repo without ref (branch/commit) used as a static file. Please specify a ref so changes can be seen.");
                }

                oldRef, exists, err := util.CachePut(cachePath, common.FILESPEC_GIT_PREFIX + url, ref);
                if (err != nil) {
                    return false, err;
                }

                if (!exists || (oldRef != ref)) {
                    gitChanges = true;
                }
            default:
                return false, fmt.Errorf("Unknown filespec type '%s': '%v'.", filespec, filespec.GetType());
        }
    }

    pathChanges, err := util.CheckFileChanges(fileCachePath, paths, quick);
    if (err != nil) {
        return false, err;
    }

    return (gitChanges || pathChanges), nil;
}
