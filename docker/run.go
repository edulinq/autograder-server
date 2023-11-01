package docker

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/mount"
    "github.com/docker/docker/pkg/stdcopy"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/util"
)

func RunContainer(imageName string, inputDir string, outputDir string, gradingID string) (string, error) {
    ctx, docker, err := getDockerClient();
    if (err != nil) {
        return "", err;
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
        return "", fmt.Errorf("Failed to create container '%s': '%w'.", name, err);
    }

    err = docker.ContainerStart(ctx, containerInstance.ID, types.ContainerStartOptions{});
    if (err != nil) {
        return "", fmt.Errorf("Failed to start container '%s' (%s): '%w'.", name, containerInstance.ID, err);
    }

    // Get the output reader before the container dies.
    out, err := docker.ContainerLogs(ctx, containerInstance.ID, types.ContainerLogsOptions{
        ShowStdout: true,
        ShowStderr: true,
        Follow: true,
    })

    if (err != nil) {
        log.Warn().Err(err).Str("container-name", name).Str("container-id", containerInstance.ID).Msg("Failed to get output from container (but run did not throw an error).");
        out = nil;
    }
    defer out.Close()

    statusChan, errorChan := docker.ContainerWait(ctx, containerInstance.ID, container.WaitConditionNotRunning);
    select {
        case err := <-errorChan:
            if (err != nil) {
                return "", fmt.Errorf("Got an error when running container '%s' (%s): '%w'.", name, containerInstance.ID, err);
            }
        case <-statusChan:
            // Waiting is complete.
    }

    output := "";

    // Read the output after the container is done.
    if (out != nil) {
        outBuffer := new(strings.Builder);
        errBuffer := new(strings.Builder);

        stdcopy.StdCopy(outBuffer, errBuffer, out);

        log.Debug().Str("container-name", name).Str("container-id", containerInstance.ID).Str("stdout", outBuffer.String()).Str("stderr", errBuffer.String()).Msg("Container output.");

        output = fmt.Sprintf("\n--- stdout ---\n%s\n--- stderr ---\n%s\n---\n", outBuffer.String(), errBuffer.String());
    }

    return output, nil;
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
