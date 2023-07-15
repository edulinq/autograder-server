package model

import (
	"fmt"
    "io/ioutil"
    "path/filepath"
    "regexp"
    "strings"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/util"
)

const COURSE_CONFIG_FILENAME = "course.json"
const DOCKER_WORK_DIR = "/autograder/work"

type DockerImageConfig struct {
    ParentName string `json:"parent-image"`
    Args []string `json:"args"`
    StaticFiles []string `json:"static-files"`
    BuildCommands []string `json:"build-commands"`
}

type AssignmentConfig struct {
    ID string  `json:"id"`
    DisplayName string `json:"display-name"`
    Files []string `json:"files"`
    Image DockerImageConfig `json:"image"`

    sourcePath string
    course *CourseConfig
}

type CourseConfig struct {
    ID string  `json:"id"`
    DisplayName string `json:"display-name"`
}

func MustLoadAssignmentConfig(path string) *AssignmentConfig {
    config, err := LoadAssignmentConfig(path);
    if (err != nil) {
        log.Fatal().Str("path", path).Err(err).Msg("Failed to load assignment config.");
    }

    return config;
}

func LoadAssignmentConfig(path string) (*AssignmentConfig, error) {
    var config AssignmentConfig;
    err := util.JSONFromFile(path, &config);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load assignment config (%s): '%w'.", path, err);
    }

    config.sourcePath, err = filepath.Abs(path);
    if (err != nil) {
        return nil, fmt.Errorf("Could not create abs path from '%s': '%w'.", path, err);
    }

    config.course, err = loadCourseConfig(filepath.Dir(path));
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config for '%s': '%w'.", path, err);
    }

    err = config.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate config (%s): '%w'.", path, err);
    }

    return &config, nil;
}

// Write a full docker build context (Dockerfile and static files) to the given directory.
func (this *AssignmentConfig) WriteDockerContext(dir string) error {
    dockerfilePath := filepath.Join(dir, "Dockerfile");
    err := this.WriteDockerfile(dockerfilePath)
    if (err != nil) {
        return err;
    }

    // The directory containing the assignment config and base for all relative paths.
    sourceDir := filepath.Dir(this.sourcePath);

    for _, relpath := range this.Image.StaticFiles {
        sourcePath := filepath.Join(sourceDir, relpath);
        destPath := filepath.Join(dir, relpath);

        err = util.CopyFile(sourcePath, destPath);
        if (err != nil) {
            return fmt.Errorf("Failed to copy static file for docker context for assignment (%s): '%w'.", this.ID, err);
        }
    }

    return nil;
}

func (this *AssignmentConfig) WriteDockerfile(path string) error {
    contents, err := this.ToDockerfile()
    if (err != nil) {
        return fmt.Errorf("Failed get contenets for docerkfile ('%s'): '%w'.", path, err);
    }

    err = ioutil.WriteFile(path, []byte(contents), 0644);
    if (err != nil) {
        return fmt.Errorf("Failed write docerkfile ('%s'): '%w'.", path, err);
    }

    return nil;
}

func (this *AssignmentConfig) ToDockerfile() (string, error) {
    // Note that we will insert blank lines for formatting.
    lines := make([]string, 0);

    lines = append(lines, fmt.Sprintf("FROM %s", this.Image.ParentName), "")

    for _, path := range this.Image.StaticFiles {
        if (filepath.IsAbs(path)) {
            return "", fmt.Errorf("All paths in an assignment config (%s) must be relative (to the assignment config file), found: '%s'.", this.sourcePath, path);
        }

        path = filepath.Clean(path);
        outPath := filepath.Join(DOCKER_WORK_DIR, path)

        lines = append(lines, fmt.Sprintf("COPY %s %s", util.DockerfilePathQuote(path), util.DockerfilePathQuote(outPath)));
    }

    // TODO(eriq): All the other stuff (copy, commands, etc).

    return strings.Join(lines, "\n"), nil;
}

func (this *AssignmentConfig) FullID() string {
    return fmt.Sprintf("%s-%s", this.course.ID, this.ID);
}

func (this *AssignmentConfig) ImageName() string {
    return strings.ToLower(fmt.Sprintf("autograder.%s.%s", this.course.ID, this.ID));
}

func (this *AssignmentConfig) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = validateID(this.ID);
    if (err != nil) {
        return err;
    }

    if (this.sourcePath == "") {
        return fmt.Errorf("Source path must not be empty.")
    }

    if (this.course == nil) {
        return fmt.Errorf("No course found for assignment.")
    }

    return nil;
}

func (this *CourseConfig) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = validateID(this.ID);
    if (err != nil) {
        return err;
    }

    return nil;
}

// Check this directory and all parent directories for a course config file.
func loadCourseConfig(basepath string) (*CourseConfig, error) {
    basepath, err := filepath.Abs(basepath);
    if (err != nil) {
        return nil, err;
    }

    for ; ; {
        configPath := filepath.Join(basepath, COURSE_CONFIG_FILENAME);

        if (!util.PathExists(configPath)) {
            // Move up to the parent.
            oldBasepath := basepath;
            basepath = filepath.Dir(basepath);

            if (oldBasepath == basepath) {
                break;
            }

            continue;
        }

        var config CourseConfig;
        err := util.JSONFromFile(configPath, &config);
        if (err != nil) {
            return nil, fmt.Errorf("Could not load course config (%s): '%w'.", configPath, err);
        }

        err = config.Validate();
        if (err != nil) {
            return nil, fmt.Errorf("Could not validate course config (%s): '%w'.", configPath, err);
        }

        return &config, nil;
    }

    return nil, fmt.Errorf("Could not locate course config.");
}

// Return a cleaned ID, or an error if the ID cannot be cleaned.
func validateID(id string) (string, error) {
    id = strings.ToLower(id);

    if (!regexp.MustCompile(`^[a-z0-9\._\-]+$`).MatchString(id)) {
        return "", fmt.Errorf("IDs must only have letters, digits, and single sequences of periods, underscores, and hyphens, found '%s'.", id);
    }

    if (regexp.MustCompile(`(^[\._\-])|(^[\._\-])$`).MatchString(id)) {
        return "", fmt.Errorf("IDs cannot start or end with periods, underscores, or hyphens, found '%s'.", id);
    }

    return id, nil;
}
