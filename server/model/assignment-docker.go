package model

// TEST
/* TEST

import (
	"context"
	"fmt"
    "io"
    "io/ioutil"
	"os"
    "path/filepath"
    "strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
    "github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"

    "github.com/eriq-augustine/autograder/util"
)


// TEST
func (this *Assignment) RunDockerGrader(submissionPath string, outputDir string) (*GradedAssignment, error) {
// func RunDockerGrader(assignment *Assignment, submissionPath string, outputDir string) (*GradedAssignment, error) {
    if (!util.PathExists(outputDir)) {
        os.MkdirAll(outputDir, 0755);
    }

    if (!util.IsEmptyDir(outputDir)) {
        return nil, fmt.Errorf("Output dir for grader is not empty.");
    }

    err := this.runGraderContainer(submissionPath, outputDir);
    if (err != nil) {
        return nil, err;
    }

    resultPath := filepath.Join(outputDir, GRADER_OUTPUT_RESULT_FILENAME);
    if (!util.PathExists(resultPath)) {
        return nil, fmt.Errorf("Cannot find output file ('%s') after grading container was run.", resultPath);
    }

    var result GradedAssignment;
    err = util.JSONFromFile(resultPath, &result);
    if (err != nil) {
        return nil, err;
    }

    return &result, nil;
}

// TODO(eriq): More gracefull errors (no panics),
// and try to diferentiate a Docker error, grader error, and submission fail.
func (this *Assignment) runGraderContainer(submissionPath string, outputDir string) error {
	ctx, docker, err := getDockerClient();
    if (err != nil) {
        return err;
    }
	defer docker.Close()

    submissionPath = util.MustAbs(submissionPath);
    outputDir = util.MustAbs(outputDir);

    // TODO(eriq): Unique name.
    name := ""

	resp, err := docker.ContainerCreate(
        ctx,
        &container.Config{
            Image: this.ImageName(),
            Tty: false,
            NetworkDisabled: true,
        },
        &container.HostConfig{
            Mounts: []mount.Mount{
                mount.Mount{
                    Type: "bind",
                    Source: submissionPath,
                    Target: "/autograder/submission",
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

*/
