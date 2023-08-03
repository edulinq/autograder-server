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
    ID string  `json:"id"`
    DisplayName string `json:"display-name"`
    Files []string `json:"files"`
    Image DockerImageConfig `json:"image"`

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

    courseConfig.Assignments[config.ID] = &config;

    return &config, nil;
}

func (this *Assignment) FullID() string {
    return fmt.Sprintf("%s-%s", this.Course.ID, this.ID);
}

func (this *Assignment) ImageName() string {
    return strings.ToLower(fmt.Sprintf("autograder.%s.%s", this.Course.ID, this.ID));
}

func (this *Assignment) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = validateID(this.ID);
    if (err != nil) {
        return err;
    }

    if (this.SourcePath == "") {
        return fmt.Errorf("Source path must not be empty.")
    }

    if (this.Course == nil) {
        return fmt.Errorf("No course found for assignment.")
    }

    return nil;
}

// Ensure the assignment is ready for grading.
func (this *Assignment) Init() error {
    return this.BuildDockerImage();
}

func MustLoadAssignmentConfig(path string) *Assignment {
    config, err := LoadAssignmentConfig(path, nil);
    if (err != nil) {
        log.Fatal().Str("path", path).Err(err).Msg("Failed to load assignment config.");
    }

    return config;
}
