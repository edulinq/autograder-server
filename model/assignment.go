package model

import (
    "errors"
    "fmt"
    "path/filepath"
    "strings"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/util"
)

const ASSIGNMENT_CONFIG_FILENAME = "assignment.json"
const DEFAULT_SUBMISSIONS_DIR = "_submissions"

const FILE_CACHE_FILENAME = "filecache.json"
const CACHE_FILENAME = "cache.json"

const CACHE_KEY_BUILD_SUCCESS = "image-build-success"

type Assignment struct {
    ID string `json:"id"`
    DisplayName string `json:"display-name"`
    SortID string `json:"sort-id"`

    CanvasID string `json:"canvas-id",omitempty`
    LatePolicy LateGradingPolicy `json:"late-policy,omitempty"`

    docker.ImageInfo

    // Ignore these fields in JSON.
    SourcePath string `json:"-"`
    Course *Course `json:"-"`
}

// Load an assignment config from a given JSON path.
// If the course config is nil, search all parent directories for the course config.
func LoadAssignmentConfig(path string, courseConfig *Course) (*Assignment, error) {
    var config Assignment;
    err := util.JSONFromFile(path, &config);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load assignment config (%s): '%w'.", path, err);
    }

    config.SourcePath = util.MustAbs(path);

    if (courseConfig == nil) {
        courseConfig, err = loadParentCourseConfig(filepath.Dir(path));
        if (err != nil) {
            return nil, fmt.Errorf("Could not load course config for '%s': '%w'.", path, err);
        }
    }
    config.Course = courseConfig;

    err = config.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate config (%s): '%w'.", path, err);
    }

    otherAssignment := courseConfig.Assignments[config.ID];
    if (otherAssignment != nil) {
        return nil, fmt.Errorf(
                "Found multiple assignments with the same ID ('%s'): ['%s', '%s'].",
                config.ID, otherAssignment.SourcePath, config.SourcePath);
    }
    courseConfig.Assignments[config.ID] = &config;

    return &config, nil;
}

func MustLoadAssignmentConfig(path string) *Assignment {
    config, err := LoadAssignmentConfig(path, nil);
    if (err != nil) {
        log.Fatal().Str("path", path).Err(err).Msg("Failed to load assignment config.");
    }

    return config;
}

func (this *Assignment) FullID() string {
    return fmt.Sprintf("%s-%s", this.Course.ID, this.ID);
}

func (this *Assignment) ImageName() string {
    return strings.ToLower(fmt.Sprintf("autograder.%s.%s", this.Course.ID, this.ID));
}

func (this *Assignment) GetImageInfo() *docker.ImageInfo {
    return &this.ImageInfo;
}

// Ensure that the assignment is formatted correctly.
// Missing optional components will be defaulted correctly.
func (this *Assignment) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = common.ValidateID(this.ID);
    if (err != nil) {
        return err;
    }

    err = this.LatePolicy.Validate();
    if (err != nil) {
        return fmt.Errorf("Failed to validate late policy: '%w'.", err);
    }

    if (this.PreStaticDockerCommands == nil) {
        this.PreStaticDockerCommands = make([]string, 0);
    }

    if (this.PostStaticDockerCommands == nil) {
        this.PostStaticDockerCommands = make([]string, 0);
    }

    if (this.StaticFiles == nil) {
        this.StaticFiles = make([]common.FileSpec, 0);
    }

    for _, staticFile := range this.StaticFiles {
        if (staticFile.IsAbs()) {
            return fmt.Errorf("All static file paths must be relative (to the assignment config file), found: '%s'.", staticFile);
        }
    }

    if (this.PreStaticFileOperations == nil) {
        this.PreStaticFileOperations = make([][]string, 0);
    }

    if (this.PostStaticFileOperations == nil) {
        this.PostStaticFileOperations = make([][]string, 0);
    }

    if (this.PostSubmissionFileOperations == nil) {
        this.PostSubmissionFileOperations = make([][]string, 0);
    }

    if (this.SourcePath == "") {
        return fmt.Errorf("Source path must not be empty.")
    }

    if (this.Course == nil) {
        return fmt.Errorf("No course found for assignment.")
    }

    if ((this.Image == "") && ((this.Invocation == nil) || (len(this.Invocation) == 0))) {
        return fmt.Errorf("Assignment image and invocation cannot both be empty.");
    }

    this.ImageInfo.Name = this.ImageName();
    this.ImageInfo.BaseDir = filepath.Dir(this.SourcePath);

    return nil;
}

func (this *Assignment) GetCacheDir() string {
    dir := filepath.Join(this.Course.GetCacheDir(), "assignment_" + this.ID);
    util.MkDir(dir);
    return dir;
}

func (this *Assignment) GetCachePath() string {
    return filepath.Join(this.GetCacheDir(), CACHE_FILENAME);
}

func (this *Assignment) GetFileCachePath() string {
    return filepath.Join(this.GetCacheDir(), FILE_CACHE_FILENAME);
}

func CompareAssignments(a *Assignment, b *Assignment) int {
    if (a == b) {
        return 0;
    }

    // Favor non-nil over nil.
    if (a == nil) {
        return 1;
    } else if (b == nil) {
        return -1;
    }

    // If both assignments have a sort key, use that for comparison.
    if ((a.SortID != "") && (b.SortID != "")) {
        return strings.Compare(a.SortID, b.SortID);
    }

    // Favor assignments with a sort key over those without.
    if (a.SortID == "") {
        return 1;
    } else if (b.SortID == "") {
        return -1;
    }

    // If both don't have sort keys, just use the IDs.
    return strings.Compare(a.ID, b.ID);
}

func (this *Assignment) BuildImage(force bool, quick bool, options *docker.BuildOptions) error {
    if (config.DOCKER_DISABLE.GetBool()) {
        return nil;
    }

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
