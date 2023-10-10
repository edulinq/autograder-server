package model

import (
    "errors"
    "fmt"
    "path/filepath"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/util"
)

const CACHE_KEY_BUILD_SUCCESS = "image-build-success"

func (this *Assignment) BuildImageQuick() error {
    return this.BuildImage(false, true, docker.NewBuildOptions());
}

func (this *Assignment) BuildImage(force bool, quick bool, options *docker.BuildOptions) error {
    if (config.DOCKER_DISABLE.GetBool()) {
        return nil;
    }

    this.dockerLock.Lock();
    defer this.dockerLock.Unlock();

    if (force) {
        quick = false;
    }

    build, err := this.needImageRebuild(quick);
    if (err != nil) {
        return fmt.Errorf("Could not check if image needs building for assignment '%s': '%w'.", this.ID, err);
    }

    if (!force && !build) {
        // Nothing has changed, skip build.
        log.Debug().Str("assignment", this.ID).Msg("No files have changed, skipping image build.");
        return nil;
    }

    buildErr := docker.BuildImageWithOptions(this.GetImageInfo(), options);

    // Always try to store the result of cache building.
    _, _, cacheErr := util.CachePut(this.GetCachePath(), CACHE_KEY_BUILD_SUCCESS, (buildErr == nil));

    return errors.Join(buildErr, cacheErr);
}

func (this *Assignment) needImageRebuild(quick bool) (bool, error) {
    // Check if the last build failed.
    lastBuildSuccess, exists, err := util.CacheFetch(this.GetCachePath(), CACHE_KEY_BUILD_SUCCESS);
    if (err != nil) {
        return false, fmt.Errorf("Failed to fetch the last build status from cahce for assignment '%s': '%w'.", this.ID, err);
    }

    lastBuildFailed := true;
    if (exists) {
        lastBuildFailed = !(lastBuildSuccess.(bool));
    }

    if (lastBuildFailed && quick) {
        return true, nil;
    }

    // Check if the image info has changed.
    imageInfoHash, err := util.MD5StringHex(util.MustToJSON(this.GetImageInfo()));
    if (err != nil) {
        return false, fmt.Errorf("Failed to hash image info for assignment '%s': '%w'.", this.ID, err);
    }

    oldHash, _, err := util.CachePut(this.GetCachePath(), "image-info", imageInfoHash);
    if (err != nil) {
        return false, fmt.Errorf("Failed to put image info hash into cahce for assignment '%s': '%w'.", this.ID, err);
    }

    imageInfoHashHasChanges := (imageInfoHash != oldHash);
    if (imageInfoHashHasChanges && quick) {
        return true, nil;
    }

    // Check if the static files have changes.
    staticFilesHaveChanges, err := this.CheckFileChanges(quick);
    if (err != nil) {
        return false, fmt.Errorf("Could not check if static files changed for assignment '%s': '%w'.", this.ID, err);
    }

    return (lastBuildFailed || imageInfoHashHasChanges || staticFilesHaveChanges), nil;
}

// Check if the assignment's static files have changes since the last time they were cached.
// This is thread-safe.
func (this *Assignment) CheckFileChanges(quick bool) (bool, error) {
    baseDir := filepath.Dir(this.SourcePath);

    fileCachePath := this.GetFileCachePath();
    cachePath := this.GetCachePath();

    paths := make([]string, 0, len(this.StaticFiles));
    gitChanges := false;

    for _, filespec := range this.StaticFiles {
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
                    log.Warn().Str("assignment", this.ID).Str("repo", url).
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
