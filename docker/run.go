package docker

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/mount"
    "github.com/docker/docker/pkg/stdcopy"

    "github.com/edulinq/autograder/common"
    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/util"
)

func RunContainer(logId log.Loggable, imageName string, inputDir string, outputDir string, gradingID string) (string, string, error) {
    ctx, docker, err := getDockerClient();
    if (err != nil) {
        return "", "", err;
    }
    defer docker.Close()

    inputDir = util.ShouldAbs(inputDir);
    outputDir = util.ShouldAbs(outputDir);

    name := cleanContainerName(fmt.Sprintf("%s-%s", gradingID, util.UUID()));

    containerInstance, err := docker.ContainerCreate(
        ctx,
        &container.Config{
            Image: imageName,
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

    if (err != nil) {
        return "", "", fmt.Errorf("Failed to create container '%s': '%w'.", name, err);
    }

    err = docker.ContainerStart(ctx, containerInstance.ID, types.ContainerStartOptions{});
    if (err != nil) {
        return "", "", fmt.Errorf("Failed to start container '%s' (%s): '%w'.", name, containerInstance.ID, err);
    }

    // Get the output reader before the container dies.
    out, err := docker.ContainerLogs(ctx, containerInstance.ID, types.ContainerLogsOptions{
        ShowStdout: true,
        ShowStderr: true,
        Follow: true,
    })

    if (err != nil) {
        log.Warn("Failed to get output from container (but run did not throw an error).",
                err, logId,
                log.NewAttr("container-name", name), log.NewAttr("container-id", containerInstance.ID));
        out = nil;
    }
    defer out.Close()

    stdout := "";
    stderr := "";

    if (out != nil) {
        outBuffer := newFixedBuffer(config.GRADER_OUTPUT_LIMIT_KB.Get() * 1024);
        errBuffer := newFixedBuffer(config.GRADER_OUTPUT_LIMIT_KB.Get() * 1024);

        _, err = stdcopy.StdCopy(outBuffer, errBuffer, out);

        stdout = outBuffer.String();
        stderr = errBuffer.String();

        log.Debug("Container output.",
                logId,
                log.NewAttr("container-name", name),
                log.NewAttr("container-id", containerInstance.ID),
                log.NewAttr("stdout", stdout),
                log.NewAttr("stderr", stderr));

        if err != nil {
            docker.ContainerKill(ctx, containerInstance.ID, "KILL");
            return stdout, stderr, err;
        }
    }

    statusChan, errorChan := docker.ContainerWait(ctx, containerInstance.ID, container.WaitConditionNotRunning);
    select {
        case err := <-errorChan:
            if (err != nil) {
                return "", "", fmt.Errorf("Got an error when running container '%s' (%s): '%w'.", name, containerInstance.ID, err);
            }
        case <-statusChan:
            // Waiting is complete.
    }

    return stdout, stderr, nil;
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

type fixedBuffer struct {
    buf *strings.Builder
    limit int
}

func newFixedBuffer(limit int) *fixedBuffer {
    buf := new(strings.Builder)
    return &fixedBuffer{
        buf: buf,
        limit: limit,
    }
}

func (this *fixedBuffer) Write(p []byte) (int, error) {
    if ((this.limit > 0) && ((this.buf.Len() + len(p)) > this.limit)) {
        return 0, &common.SecureError{
            "Output exceeds limit. Do you have an infinite loop?",
        }
    }
    return this.buf.Write(p);
}

func (this *fixedBuffer) String() string {
    return this.buf.String();
}
