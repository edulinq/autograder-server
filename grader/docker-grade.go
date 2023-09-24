package grader

// Handle running docker containers for grading.

import (
	"fmt"
	"os"
    "path/filepath"
    "regexp"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// Grade using a docker container.
// Directory information:
//  - input -- A temp dir that will be mounted at DOCKER_INPUT_DIR (read-only).
//  - output -- Passed in directory that will be mounted at DOCKER_OUTPUT_DIR.
//  - work -- Should already be created inside the docker image, will only exist within the container.
func RunDockerGrader(assignment *model.Assignment, submissionPath string, outputDir string, options GradeOptions, gradingID string) (*model.GradedAssignment, error) {
    os.MkdirAll(outputDir, 0755);
    if (!util.IsEmptyDir(outputDir)) {
        return nil, fmt.Errorf("Output dir for docker grader is not empty.");
    }

    // Create a temp directory to use for input (will be mounted to the container).
    tempInputDir, err := os.MkdirTemp("", "autograding-docker-input-");
    if (err != nil) {
        return nil, fmt.Errorf("Could not create temp input dir: '%w'.", err);
    }

    if (options.LeaveTempDir) {
        log.Info().Str("path", tempInputDir).Msg("Leaving behind temp input dir.");
    } else {
        defer os.RemoveAll(tempInputDir);
    }

    // Copy over submission files to the temp input dir.
    err = util.CopyDirent(submissionPath, tempInputDir, true);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to copy over submission/input contents: '%w'.", err);
    }

    err = runGraderContainer(assignment, tempInputDir, outputDir, gradingID);
    if (err != nil) {
        return nil, err;
    }

    resultPath := filepath.Join(outputDir, model.GRADER_OUTPUT_RESULT_FILENAME);
    if (!util.PathExists(resultPath)) {
        return nil, fmt.Errorf("Cannot find output file ('%s') after the grading container (%s) was run.", resultPath, assignment.ImageName());
    }

    var result model.GradedAssignment;
    err = util.JSONFromFile(resultPath, &result);
    if (err != nil) {
        return nil, err;
    }

    return &result, nil;
}

func runGraderContainer(assignment *model.Assignment, inputDir string, outputDir string, gradingID string) error {
	ctx, docker, err := getDockerClient();
    if (err != nil) {
        return err;
    }
	defer docker.Close()

    inputDir = util.MustAbs(inputDir);
    outputDir = util.MustAbs(outputDir);

    name := cleanContainerName(fmt.Sprintf("%s-%s", gradingID, util.UUID()));

	resp, err := docker.ContainerCreate(
        ctx,
        &container.Config{
            Image: assignment.ImageName(),
            Tty: false,
            NetworkDisabled: true,
        },
        &container.HostConfig{
            AutoRemove: true,
            Mounts: []mount.Mount{
                mount.Mount{
                    Type: "bind",
                    Source: inputDir,
                    Target: "/autograder/input",
                    ReadOnly: true,
                },
                mount.Mount{
                    Type: "bind",
                    Source: outputDir,
                    Target: "/autograder/output",
                    ReadOnly: false,
                },
            },
        },
	    nil,
        nil,
        name)

	if err != nil {
		panic(err)
	}

	if err := docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := docker.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := docker.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

    return nil;
}

func cleanContainerName(text string) string {
    pattern := regexp.MustCompile(`[^a-zA-Z0-9_\.\-]`);
    text = pattern.ReplaceAllString(text, "");

    match, _ := regexp.MatchString(`^[a-zA-Z0-9]`, text);
    if (!match) {
        text = "a" + text;
    }

    return text;
}
