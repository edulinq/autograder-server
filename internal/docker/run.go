package docker

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

// Run a container.
// Returns: (stdout, stderr, timeout?, error)
func RunContainer(logId log.Loggable, imageName string, inputDir string, outputDir string, baseID string, maxRuntimeSecs int) (string, string, bool, error) {
	ctx, docker, err := getDockerClient()
	if err != nil {
		return "", "", false, err
	}
	defer docker.Close()

	inputDir = util.ShouldAbs(inputDir)
	outputDir = util.ShouldAbs(outputDir)

	name := cleanContainerName(fmt.Sprintf("%s-%s", baseID, util.UUID()))

	containerInstance, err := docker.ContainerCreate(
		ctx,
		&container.Config{
			Image:           imageName,
			Tty:             false,
			NetworkDisabled: true,
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				mount.Mount{
					Type:     "bind",
					Source:   inputDir,
					Target:   "/autograder/input",
					ReadOnly: true,
				},
				mount.Mount{
					Type:     "bind",
					Source:   outputDir,
					Target:   "/autograder/output",
					ReadOnly: false,
				},
			},
		},
		nil,
		nil,
		name)

	if err != nil {
		return "", "", false, fmt.Errorf("Failed to create container '%s': '%w'.", name, err)
	}

	defer func() {
		// The container should have already gracefull exited.
		// If not, kill it without any grace.
		// Ignore any errors.
		docker.ContainerKill(ctx, containerInstance.ID, "KILL")

		err = docker.ContainerRemove(ctx, containerInstance.ID, container.RemoveOptions{
			Force: true,
		})
		if err != nil {
			log.Warn("Failed to remove container.",
				err, logId,
				log.NewAttr("container-name", name), log.NewAttr("container-id", containerInstance.ID))
		}
	}()

	err = docker.ContainerStart(ctx, containerInstance.ID, container.StartOptions{})
	if err != nil {
		return "", "", false, fmt.Errorf("Failed to start container '%s' (%s): '%w'.", name, containerInstance.ID, err)
	}

	// Get the output reader before the container dies.
	out, err := docker.ContainerLogs(ctx, containerInstance.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})

	if err != nil {
		log.Warn("Failed to get output from container (but run did not throw an error).",
			err, logId,
			log.NewAttr("container-name", name), log.NewAttr("container-id", containerInstance.ID))
		out = nil
	} else {
		defer out.Close()
	}

	timeout := false
	timeoutContext := ctx
	var cancel context.CancelFunc
	if maxRuntimeSecs > 0 {
		timeoutContext, cancel = context.WithTimeout(ctx, time.Duration(maxRuntimeSecs)*time.Second)
		defer cancel()
	}

	statusChan, errorChan := docker.ContainerWait(timeoutContext, containerInstance.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errorChan:
		if err != nil {
			// On a timeout, kill the container without any grace, exit this select, and try to recover the output.
			if err.Error() == "context deadline exceeded" {
				docker.ContainerKill(ctx, containerInstance.ID, "KILL")
				timeout = true
				break
			}

			return "", "", false, fmt.Errorf("Got an error when running container '%s' (%s): '%w'.", name, containerInstance.ID, err)
		}
	case <-statusChan:
		// Waiting is complete.
	}

	stdout := ""
	stderr := ""

	// Read the output after the container is done.
	if out != nil {
		outBuffer := new(strings.Builder)
		errBuffer := new(strings.Builder)

		stdcopy.StdCopy(outBuffer, errBuffer, out)

		stdout = outBuffer.String()
		stderr = errBuffer.String()

		log.Debug("Container output.",
			logId,
			log.NewAttr("container-name", name),
			log.NewAttr("container-id", containerInstance.ID),
			log.NewAttr("stdout", stdout),
			log.NewAttr("stderr", stderr))
	}

	return stdout, stderr, timeout, nil
}

func cleanContainerName(text string) string {
	pattern := regexp.MustCompile(`[^a-zA-Z0-9_\.\-]`)
	text = pattern.ReplaceAllString(text, "")

	match, _ := regexp.MatchString(`^[a-zA-Z0-9]`, text)
	if !match {
		text = "a" + text
	}

	return text
}
