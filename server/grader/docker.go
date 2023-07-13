package grader

import (
	"context"
	"fmt"
    "io"
    "path/filepath"
	"os"
    "strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
    "github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"

    "github.com/eriq-augustine/autograder/model"
)

func getDockerClient() (context.Context, *client.Client, error) {
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return ctx, nil, fmt.Errorf("Cannot create Docker client: '%w'.", err);
	}

    return ctx, docker, nil;
}

func RunContainerGrader(imageName string, submissionPath string) error {
	ctx, docker, err := getDockerClient();
    if (err != nil) {
        return err;
    }
	defer docker.Close()

    submissionPath, err = filepath.Abs(submissionPath);
    if (err != nil) {
        return fmt.Errorf("Could not create abs path for submission mount from '%s': '%w'.", submissionPath, err);
    }

    // TODO(eriq): Unique name.
    name := ""

	resp, err := docker.ContainerCreate(
        ctx,
        &container.Config{
            Image: imageName,
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

	out, err := docker.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

    return nil;
}

func BuildAssignmentImage(courseID string, assignment *model.AssignmentConfig) (string, error) {
    // TODO(eriq): Sanitization should be done earlier on IDs.
    imageName := strings.ToLower(fmt.Sprintf("autograder.%s.%s", courseID, assignment.ID));

	ctx, docker, err := getDockerClient();
    if (err != nil) {
        return "", err;
    }
	defer docker.Close()

    tempDir, err := os.MkdirTemp("", "docker-build-" + imageName + "-");
    if (err != nil) {
        return "", fmt.Errorf("Failed to create temp build directory for '%s': '%w'.", imageName, err);
    }
    defer os.RemoveAll(tempDir);

    err = assignment.WriteDockerContext(tempDir);
    if (err != nil) {
        return "", err;
    }

    // TODO(eriq): Version
    buildOptions := types.ImageBuildOptions{
        Tags: []string{imageName},
        Dockerfile: "Dockerfile",
    };

    // Create the build context by adding all the relevant files.
    tar, err := archive.TarWithOptions(tempDir, &archive.TarOptions{});
    if (err != nil) {
        return "", fmt.Errorf("Failed to create tar build context for image '%s': '%w'.", imageName, err);
    }

    response, err := docker.ImageBuild(ctx, tar, buildOptions);
    if (err != nil) {
        return "", fmt.Errorf("Failed to build assignment image ('%s'): '%w'.", imageName, err);
    }
    defer response.Body.Close();

    _, err = io.Copy(os.Stdout, response.Body);
    if (err != nil) {
        return "", fmt.Errorf("Unable to get response body from build of '%s': '%w'.", imageName, err);
    }

    return imageName, nil;
}
