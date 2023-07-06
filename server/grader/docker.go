package grader

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
	"github.com/docker/docker/client"
    "github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
)

// TODO(eriq): Break up into more files.
type AssignmentConfig struct {
    ID string  `json:"id"`
    DisplayName string `json:"display-name"`
    Image DockerImageConfig `json:"image"`
}

// TODO(eriq): Break up into more files.
type DockerImageConfig struct {
    ParentName string `json:"parent"`
    Args []string `json:"args"`
    Files string `json:"files"`
    BuildCommands string `json:"build-commands"`
}

func (this DockerImageConfig) ToDockerfile() string {
    lines := make([]string, 0);

    lines = append(lines, fmt.Sprintf("FROM %s", this.ParentName));

    // TODO(eriq): All the other stuff (copy, commands, etc).

    return strings.Join(lines, "\n");
}

func getDockerClient() (context.Context, *client.Client, error) {
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return ctx, nil, fmt.Errorf("Cannot create Docker client: '%w'.", err);
	}

    return ctx, docker, nil;
}

// TEST
func RunContainerGrader(imageName string) error {
	ctx, docker, err := getDockerClient();
    if (err != nil) {
        return err;
    }
	defer docker.Close()

	resp, err := docker.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Tty:   false,
	}, nil, nil, nil, "")
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

func BuildAssignmentImage(courseID string, assignment AssignmentConfig) (string, error) {
    // TODO(eriq): Sanitization should be done earlier on IDS.
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

    dockerfilePath := filepath.Join(tempDir, "Dockerfile");

    err = ioutil.WriteFile(dockerfilePath, []byte(assignment.Image.ToDockerfile()), 0644);
    if (err != nil) {
        return "", fmt.Errorf("Failed to write dockerfile (%s) for '%s': '%w'", dockerfilePath, imageName, err);
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

    // TODO(eriq): Create a build dir, copy files to build into image, and set it as the build context.

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
