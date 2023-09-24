package grader

// Handle building docker images for grading.

import (
    "bufio"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

	"github.com/docker/docker/api/types"
    "github.com/docker/docker/pkg/archive"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const DOCKER_CONFIG_FILENAME = "config.json"

const DOCKER_BASE_DIR = "/autograder"
const DOCKER_INPUT_DIR = DOCKER_BASE_DIR + "/input"
const DOCKER_OUTPUT_DIR = DOCKER_BASE_DIR + "/output"
const DOCKER_WORK_DIR = DOCKER_BASE_DIR + "/work"
const DOCKER_CONFIG_PATH = DOCKER_BASE_DIR + "/" + DOCKER_CONFIG_FILENAME

type DockerBuildOptions struct {
    Rebuild bool `help:"Rebuild images ignoring caches." default:"false"`
}

func NewDockerBuildOptions() *DockerBuildOptions {
    return &DockerBuildOptions{
        Rebuild: false,
    };
}

func BuildDockerImagesJoinErrors(buildOptions *DockerBuildOptions) ([]string, error) {
    imageNames, errs := BuildDockerImages(buildOptions);
    return imageNames, errors.Join(errs...);
}

func BuildDockerImages(buildOptions *DockerBuildOptions) ([]string, []error) {
    errs := make([]error, 0);
    imageNames := make([]string, 0);

    for _, course := range courses {
        for _, assignment := range course.Assignments {
            err := BuildDockerImageWithOptions(assignment, buildOptions);
            if (err != nil) {
                errs = append(errs, fmt.Errorf("Failed to build docker grader image for assignment (%s): '%w'.", assignment.FullID(), err));
            } else {
                imageNames = append(imageNames, assignment.ImageName());
            }
        }
    }

    return imageNames, errs;
}

func BuildDockerImage(assignment *model.Assignment) error {
    return BuildDockerImageWithOptions(assignment, NewDockerBuildOptions());
}

func BuildDockerImageWithOptions(assignment *model.Assignment, options *DockerBuildOptions) error {
    tempDir, err := os.MkdirTemp("", "autograder-docker-build-" + assignment.ImageName() + "-");
    if (err != nil) {
        return fmt.Errorf("Failed to create temp build directory for '%s': '%w'.", assignment.ImageName(), err);
    }

    if (config.DEBUG.GetBool()) {
        log.Info().Str("path", tempDir).Msg("Leaving behind temp building dir.");
    } else {
        defer os.RemoveAll(tempDir);
    }

    err = writeDockerContext(assignment, tempDir);
    if (err != nil) {
        return err;
    }

    buildOptions := types.ImageBuildOptions{
        Tags: []string{assignment.ImageName()},
        Dockerfile: "Dockerfile",
    };

    if (options.Rebuild) {
        buildOptions.NoCache = true;
    }

    // Create the build context by adding all the relevant files.
    tar, err := archive.TarWithOptions(tempDir, &archive.TarOptions{});
    if (err != nil) {
        return fmt.Errorf("Failed to create tar build context for image '%s': '%w'.", assignment.ImageName(), err);
    }

    return buildDockerImage(buildOptions, tar);
}

func buildDockerImage(buildOptions types.ImageBuildOptions, tar io.ReadCloser) error {
	ctx, docker, err := getDockerClient();
    if (err != nil) {
        return err;
    }
	defer docker.Close()

    response, err := docker.ImageBuild(ctx, tar, buildOptions);
    if (err != nil) {
        return fmt.Errorf("Failed to run docker image build command: '%w'.", err);
    }
    defer response.Body.Close();

    buildStringOutput := strings.Builder{};

    responseScanner := bufio.NewScanner(response.Body);
    for responseScanner.Scan() {
        line := responseScanner.Text();

        line = strings.TrimSpace(line);
        if (line == "") {
            continue;
        }

        jsonData, err := util.JSONMapFromString(line);
        if (err != nil) {
            return fmt.Errorf("Docker build output line is not JSON ('%s'): '%w'.", line, err);
        }

        rawText, ok := jsonData["error"];
        if (ok) {
            text, ok := rawText.(string);
            if (!ok) {
                text = "<ERROR: Docker output JSON value is not a string.>";
            }

            return fmt.Errorf("Docker image build failed with message: '%s'.", text);
        }

        rawText, ok = jsonData["stream"];
        if (ok) {
            text, ok := rawText.(string);
            if (!ok) {
                text = "<ERROR: Docker output JSON value is not a string.>";
            }

            buildStringOutput.WriteString(text);
        }
    }

    err = responseScanner.Err();
    if (err != nil) {
        return fmt.Errorf("Failed to scan docker image build responseL '%w'.", err);
    }

    log.Debug().Str("image-build-output", buildStringOutput.String()).Msg("Image Build Output");

    return nil;
}

// Write a full docker build context (Dockerfile and static files) to the given directory.
func writeDockerContext(assignment *model.Assignment, dir string) error {
    // The directory containing the assignment config and base for all relative paths.
    sourceDir := filepath.Dir(assignment.SourcePath);

    _, _, workDir, err := createStandardGradingDirs(dir);
    if (err != nil) {
        return fmt.Errorf("Could not create standard grading directories: '%w'.", err);
    }

    // Copy over the static files (and do any file ops).
    err = copyAssignmentFiles(sourceDir, workDir, dir,
            assignment.StaticFiles, false, assignment.PreStaticFileOperations, assignment.PostStaticFileOperations);
    if (err != nil) {
        return fmt.Errorf("Failed to copy static assignment files: '%w'.", err);
    }

    dockerConfigPath := filepath.Join(dir, DOCKER_CONFIG_FILENAME);
    err = util.ToJSONFile(assignment.GetDockerAssignmentConfig(), dockerConfigPath);
    if (err != nil) {
        return fmt.Errorf("Failed to create docker config file: '%w'.", err);
    }

    dockerfilePath := filepath.Join(dir, "Dockerfile");
    err = writeDockerfile(assignment, workDir, dockerfilePath)
    if (err != nil) {
        return err;
    }

    return nil;
}

func writeDockerfile(assignment *model.Assignment, workDir string, path string) error {
    contents, err := toDockerfile(assignment, workDir)
    if (err != nil) {
        return fmt.Errorf("Failed get contenets for dockerfile ('%s'): '%w'.", path, err);
    }

    err = util.WriteFile(contents, path);
    if (err != nil) {
        return fmt.Errorf("Failed write dockerfile ('%s'): '%w'.", path, err);
    }

    return nil;
}

func toDockerfile(assignment *model.Assignment, workDir string) (string, error) {
    // Note that we will insert blank lines for formatting.
    lines := make([]string, 0);

    lines = append(lines, fmt.Sprintf("FROM %s", assignment.Image), "")

    // Ensure standard directories are created.
    lines = append(lines, "# Core directories");
    for _, dir := range []string{DOCKER_BASE_DIR, DOCKER_INPUT_DIR, DOCKER_OUTPUT_DIR, DOCKER_WORK_DIR} {
        lines = append(lines, fmt.Sprintf("RUN mkdir -p '%s'", dir));
    }
    lines = append(lines, "");

    // Set the working directory.
    lines = append(lines, fmt.Sprintf("WORKDIR %s", DOCKER_BASE_DIR), "")

    // Copy over the config file.
    lines = append(lines, fmt.Sprintf("COPY %s %s", DOCKER_CONFIG_FILENAME, DOCKER_CONFIG_PATH), "");

    // Append pre-static docker commands.
    lines = append(lines, "# Pre-Static Assignment Commands");
    lines = append(lines, assignment.PreStaticDockerCommands...);
    lines = append(lines, "");

    // Copy over all the contents of the work directory (this is after post-static file ops).
    dirents, err := os.ReadDir(workDir);
    if (err != nil) {
        return "", fmt.Errorf("Failed to list work dir ('%s') for static files: '%w'.", workDir, err);
    }

    lines = append(lines, "# Static Files");
    for _, dirent := range dirents {
        sourcePath := util.DockerfilePathQuote(filepath.Join(model.GRADING_WORK_DIRNAME, dirent.Name()));
        destPath := util.DockerfilePathQuote(filepath.Join(DOCKER_WORK_DIR, dirent.Name()));

        lines = append(lines, fmt.Sprintf("COPY %s %s", sourcePath, destPath));
    }
    lines = append(lines, "");

    // Append post-static docker commands.
    lines = append(lines, "# Post-Static Assignment Commands");
    lines = append(lines, assignment.PostStaticDockerCommands...);
    lines = append(lines, "");

    return strings.Join(lines, "\n"), nil;
}
