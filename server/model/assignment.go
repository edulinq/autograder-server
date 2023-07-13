package model

import (
	"fmt"
    "io/ioutil"
    "path/filepath"
    "strings"

    "github.com/eriq-augustine/autograder/util"
)

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
