package model

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

const DOCKER_WORK_DIR = "/autograder/work"

type DockerImageConfig struct {
    ParentName string `json:"parent-image"`
    Args []string `json:"args"`
    StaticFiles []string `json:"static-files"`
    BuildCommands []string `json:"build-commands"`
}

func getDockerClient() (context.Context, *client.Client, error) {
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return ctx, nil, fmt.Errorf("Cannot create Docker client: '%w'.", err);
	}

    return ctx, docker, nil;
}

func (this *Assignment) RunGrader(submissionPath string) error {
	ctx, docker, err := getDockerClient();
    if (err != nil) {
        return err;
    }
	defer docker.Close()

    submissionPath = util.MustAbs(submissionPath);

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

func (this *Assignment) BuildDockerImage() error {
    imageName := this.ImageName();

	ctx, docker, err := getDockerClient();
    if (err != nil) {
        return err;
    }
	defer docker.Close()

    tempDir, err := os.MkdirTemp("", "docker-build-" + imageName + "-");
    if (err != nil) {
        return fmt.Errorf("Failed to create temp build directory for '%s': '%w'.", imageName, err);
    }
    defer os.RemoveAll(tempDir);

    err = this.WriteDockerContext(tempDir);
    if (err != nil) {
        return err;
    }

    // TODO(eriq): Version
    buildOptions := types.ImageBuildOptions{
        Tags: []string{imageName},
        Dockerfile: "Dockerfile",
    };

    // Create the build context by adding all the relevant files.
    tar, err := archive.TarWithOptions(tempDir, &archive.TarOptions{});
    if (err != nil) {
        return fmt.Errorf("Failed to create tar build context for image '%s': '%w'.", imageName, err);
    }

    response, err := docker.ImageBuild(ctx, tar, buildOptions);
    if (err != nil) {
        return fmt.Errorf("Failed to build assignment image ('%s'): '%w'.", imageName, err);
    }
    defer response.Body.Close();

    _, err = io.Copy(os.Stdout, response.Body);
    if (err != nil) {
        return fmt.Errorf("Unable to get response body from build of '%s': '%w'.", imageName, err);
    }

    return nil;
}

// Write a full docker build context (Dockerfile and static files) to the given directory.
func (this *Assignment) WriteDockerContext(dir string) error {
    dockerfilePath := filepath.Join(dir, "Dockerfile");
    err := this.WriteDockerfile(dockerfilePath)
    if (err != nil) {
        return err;
    }

    // The directory containing the assignment config and base for all relative paths.
    sourceDir := filepath.Dir(this.SourcePath);

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

func (this *Assignment) WriteDockerfile(path string) error {
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

func (this *Assignment) ToDockerfile() (string, error) {
    // Note that we will insert blank lines for formatting.
    lines := make([]string, 0);

    lines = append(lines, fmt.Sprintf("FROM %s", this.Image.ParentName), "")

    for _, path := range this.Image.StaticFiles {
        if (filepath.IsAbs(path)) {
            return "", fmt.Errorf("All paths in an assignment config (%s) must be relative (to the assignment config file), found: '%s'.", this.SourcePath, path);
        }

        path = filepath.Clean(path);
        outPath := filepath.Join(DOCKER_WORK_DIR, path)

        lines = append(lines, fmt.Sprintf("COPY %s %s", util.DockerfilePathQuote(path), util.DockerfilePathQuote(outPath)));
    }

    // TODO(eriq): All the other stuff (copy, commands, etc).

    return strings.Join(lines, "\n"), nil;
}
