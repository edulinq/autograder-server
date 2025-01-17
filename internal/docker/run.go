package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

type containerOutput struct {
	Stdout    string
	Stderr    string
	Truncated bool
	Err       error
}

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
			LogConfig: container.LogConfig{
				// Don't store any logs, we will copy stdout/stderr directly.
				Type: "none",
			},
		},
		nil,
		nil,
		name)

	if err != nil {
		return "", "", false, fmt.Errorf("Failed to create container '%s': '%w'.", name, err)
	}

	// Ensure the container is removed.
	defer func() {
		// The container should have already gracefully exited.
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

	// Attach to the container so we can get stdout and stderr.
	connection, err := docker.ContainerAttach(ctx, containerInstance.ID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return "", "", false, fmt.Errorf("Failed to attach to container '%s' (%s): '%w'.", name, containerInstance.ID, err)
	}
	defer connection.Conn.Close()

	// Handle copying (and possibly truncating) stdout/stderr.
	outputWaitGroup := &sync.WaitGroup{}
	outputWaitGroup.Add(1)
	output := &containerOutput{}
	go handleContainerOutput(output, outputWaitGroup, connection.Reader)

	err = docker.ContainerStart(ctx, containerInstance.ID, container.StartOptions{})
	if err != nil {
		return "", "", false, fmt.Errorf("Failed to start container '%s' (%s): '%w'.", name, containerInstance.ID, err)
	}

	// Set a timeout for the container.
	timeout := false
	timeoutContext := ctx
	var cancel context.CancelFunc
	if maxRuntimeSecs > 0 {
		timeoutContext, cancel = context.WithTimeout(timeoutContext, time.Duration(maxRuntimeSecs)*time.Second)
		defer cancel()
	}

	// wait for the container to finish.
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

	// Wait for output to get copied.
	outputWaitGroup.Wait()

	log.Debug("Container output.",
		logId,
		log.NewAttr("container-name", name),
		log.NewAttr("container-id", containerInstance.ID),
		log.NewAttr("stdout", output.Stdout),
		log.NewAttr("stderr", output.Stderr),
		log.NewAttr("output-truncated", output.Truncated),
		output.Err,
	)

	return output.Stdout, output.Stderr, timeout, nil
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

// Read a maximum amount from the container's stdout/stderr, parse the two from the common stream, and signal completion.
func handleContainerOutput(output *containerOutput, outputWaitGroup *sync.WaitGroup, containerStream io.Reader) {
	defer outputWaitGroup.Done()

	maxSizeKB := config.DOCKER_MAX_OUTPUT_SIZE_KB.Get()
	bufferLen := maxSizeKB * 1024

	// Make the first full (or short) read.
	buffer := make([]byte, bufferLen)
	_, err := io.ReadFull(containerStream, buffer)
	if (err != nil) && (err != io.EOF) && (err != io.ErrUnexpectedEOF) {
		output.Err = fmt.Errorf("Failed to read container output into temporary buffer: '%w'.", err)
		return
	}

	// Check for too much output.
	overflowBuffer := make([]byte, 1)
	readCount, _ := containerStream.Read(overflowBuffer)
	if readCount == 1 {
		output.Truncated = true
	}

	// Parse stdout and stderr out of the output stream.
	outBuffer := new(strings.Builder)
	errBuffer := new(strings.Builder)

	stdcopy.StdCopy(outBuffer, errBuffer, bytes.NewReader(buffer))

	// Denote truncated streams.
	if output.Truncated {
		message := fmt.Sprintf("\n\nCombined output (stdout + stderr) exceeds maximum size (%d KB), output has been truncated.", maxSizeKB)

		if outBuffer.Len() > 0 {
			outBuffer.WriteString(message)
		}

		if errBuffer.Len() > 0 {
			errBuffer.WriteString(message)
		}
	}

	output.Stdout = outBuffer.String()
	output.Stderr = errBuffer.String()
}
