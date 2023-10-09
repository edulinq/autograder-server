package model

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/util"
)

const ASSIGNMENT_CONFIG_FILENAME = "assignment.json"
const DEFAULT_SUBMISSIONS_DIR = "_submissions";

const GRADING_INPUT_DIRNAME = "input"
const GRADING_OUTPUT_DIRNAME = "output"
const GRADING_WORK_DIRNAME = "work"

const GRADER_OUTPUT_RESULT_FILENAME = "result.json"
const GRADER_OUTPUT_SUMMARY_FILENAME = "summary.json"

const FILE_CACHE_FILENAME = "filecache.json"
const CACHE_FILENAME = "cache.json"

type Assignment struct {
    ID string `json:"id"`
    DisplayName string `json:"display-name"`
    SortID string `json:"sort-id"`

    CanvasID string `json:"canvas-id",omitempty`
    LatePolicy LateGradingPolicy `json:"late-policy,omitempty"`

    Image string `json:"image,omitempty"`
    PreStaticDockerCommands []string `json:"pre-static-docker-commands,omitempty"`
    PostStaticDockerCommands []string `json:"post-static-docker-commands,omitempty"`

    Invocation []string `json:"invocation,omitempty"`

    StaticFiles []FileSpec `json:"static-files,omitempty"`
    PreStaticFileOperations [][]string `json:"pre-static-files-ops,omitempty"`
    PostStaticFileOperations [][]string `json:"post-static-files-ops,omitempty"`

    PostSubmissionFileOperations [][]string `json:"post-submission-files-ops,omitempty"`

    // Ignore these fields in JSON.
    SourcePath string `json:"-"`
    Course *Course `json:"-"`
}

// A subset of assignment that is passed to docker images for config.
type DockerAssignmentConfig struct {
    ID string `json:"id"`
    DisplayName string `json:"display-name"`

    PostSubmissionFileOperations [][]string `json:"post-submission-files-ops,omitempty"`
}

func (this *Assignment) GetDockerAssignmentConfig() *DockerAssignmentConfig {
    return &DockerAssignmentConfig{
        ID: this.ID,
        DisplayName: this.DisplayName,
        PostSubmissionFileOperations: this.PostSubmissionFileOperations,
    };
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

// Ensure that the assignment is formatted correctly.
// Missing optional components will be defaulted correctly.
func (this *Assignment) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = ValidateID(this.ID);
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
        this.StaticFiles = make([]FileSpec, 0);
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

    return nil;
}

func (this *Assignment) GetCacheDir() string {
    return filepath.Join(this.Course.GetCacheDir(), "assignment_" + this.ID);
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

// Check if the assignment's static files have changes since the last time they were cached.
// This is thread-safe.
func (this *Assignment) HaveStaticFilesChanges(quick bool) (bool, error) {
    cacheDir := this.GetCacheDir();

    err := util.MkDir(cacheDir);
    if (err != nil) {
        return false, fmt.Errorf("Unable to create cache dir '%s': '%w'.", cacheDir, err);
    }

    fileCachePath := filepath.Join(cacheDir, FILE_CACHE_FILENAME);
    cachePath := filepath.Join(cacheDir, CACHE_FILENAME);

    paths := make([]string, 0, len(this.StaticFiles));
    gitChanges := false;

    for _, filespec := range this.StaticFiles {
        if (quick && gitChanges) {
            return true, nil;
        }

        switch (filespec.GetType()) {
            case FILESPEC_TYPE_PATH:
                // Collect paths to test all at once.
                paths = append(paths, filespec.GetPath());
            case FILESPEC_TYPE_GIT:
                // Check git refs for changes.
                url, _, ref, err := filespec.ParseGitParts();
                if (err != nil) {
                    return false, err;
                }

                if (ref == "") {
                    log.Warn().Str("assignment", this.ID).Str("repo", url).
                            Msg("Git repo without ref (branch/commit) used as a static file. Please specify a ref so changes can be seen.");
                }

                oldRef, exists, err := util.CachePut(cachePath, FILESPEC_GIT_PREFIX + url, ref);
                if (err != nil) {
                    return false, err;
                }

                if (!exists || (oldRef != ref)) {
                    gitChanges = true;
                }
            default:
                return false, fmt.Errorf("Unknown filespec type '%s': '%s'.", filespec, filespec.GetType());
        }
    }

    pathChanges, err := util.HaveFilesChanges(fileCachePath, paths, quick);
    if (err != nil) {
        return false, err;
    }

    return (gitChanges || pathChanges), nil;
}
