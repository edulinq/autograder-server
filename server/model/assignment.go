package model

import (
	"fmt"
    "path/filepath"
    "strings"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/util"
)

const ASSIGNMENT_CONFIG_FILENAME = "assignment.json"

type Assignment struct {
    ID string `json:"id"`
    DisplayName string `json:"display-name"`

    Image string `json:"image,omitempty"`
    PreStaticDockerCommands []string `json:"pre-static-docker-commands,omitempty"`
    PostStaticDockerCommands []string `json:"post-static-docker-commands,omitempty"`

    Invocation []string `json:"invocation,omitempty"`

    StaticFiles []string `json:"static-files,omitempty"`
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

    if (this.PreStaticDockerCommands == nil) {
        this.PreStaticDockerCommands = make([]string, 0);
    }

    if (this.PostStaticDockerCommands == nil) {
        this.PostStaticDockerCommands = make([]string, 0);
    }

    if (this.StaticFiles == nil) {
        this.StaticFiles = make([]string, 0);
    }

    for _, staticFile := range this.StaticFiles {
        if (filepath.IsAbs(staticFile)) {
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
