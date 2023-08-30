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

type DockerBuildOptions struct {
    Rebuild bool `help:"Rebuild images ignoring caches." default:"false"`
}

func CanAccessDocker() bool {
    _, docker, err := getDockerClient();
    if (docker != nil) {
        defer docker.Close();
    }

    return (err == nil);
}

func getDockerClient() (context.Context, *client.Client, error) {
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return ctx, nil, fmt.Errorf("Cannot create Docker client: '%w'.", err);
	}

    return ctx, docker, nil;
}

func NewDockerBuildOptions() *DockerBuildOptions {
    return &DockerBuildOptions{
        Rebuild: false,
    };
}

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

func (this *Assignment) BuildDockerImage() error {
    return this.BuildDockerImageWithOptions(NewDockerBuildOptions());
}

func (this *Assignment) BuildDockerImageWithOptions(options *DockerBuildOptions) error {
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

    if (options.Rebuild) {
        buildOptions.NoCache = true;
    }

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

    for _, relpath := range this.StaticFiles {
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

    lines = append(lines, fmt.Sprintf("FROM %s", this.Image), "")

    for _, path := range this.StaticFiles {
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
