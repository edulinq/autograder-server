package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var extraInitTimeSecs int = 5

type containerOutput struct {
	Stdout    string
	Stderr    string
	Truncated bool
	Err       error
}

type MountInfo struct {
	Source   string
	Target   string
	ReadOnly bool
}

// Run a grading container.
// Returns: (stdout, stderr, timeout?, canceled?, error)
func RunGradingContainer(ctx context.Context, logId log.Loggable, imageName string, inputDir string, outputDir string, baseID string, maxRuntimeSecs int) (string, string, bool, bool, error) {
	mounts := []MountInfo{
		MountInfo{
			Source:   util.ShouldAbs(inputDir),
			Target:   "/autograder/input",
			ReadOnly: true,
		},
		MountInfo{
			Source:   util.ShouldAbs(outputDir),
			Target:   "/autograder/output",
			ReadOnly: false,
		},
	}

	return RunContainer(ctx, logId, imageName, mounts, nil, baseID, maxRuntimeSecs)
}

// Run a container.
// Returns: (stdout, stderr, timeout?, canceled?, error)
func RunContainer(ctx context.Context, logId log.Loggable, imageName string, mounts []MountInfo, cmd []string, baseID string, maxRuntimeSecs int) (string, string, bool, bool, error) {
	var stdout string
	var stderr string
	var tempTimeout bool
	var timeout bool
	var canceled bool
	var err error

	runFunc := func(softTimeoutCtx context.Context) {
		stdout, stderr, tempTimeout, err = runContainerInternal(softTimeoutCtx, logId, imageName, mounts, cmd, baseID)
		timeout = timeout || tempTimeout
	}

	if maxRuntimeSecs > 0 {
		softTimeoutMS := int64(maxRuntimeSecs * 1000)
		hardTimeoutMS := int64((maxRuntimeSecs + extraInitTimeSecs) * 1000)
		tempTimeout = !util.RunWithTimeoutFull(softTimeoutMS, hardTimeoutMS, ctx, runFunc)
		timeout = timeout || tempTimeout
	} else {
		runFunc(ctx)
	}

	timeout = timeout || errors.Is(ctx.Err(), context.DeadlineExceeded)
	canceled = errors.Is(ctx.Err(), context.Canceled)

	// Clear the error if it was caused by a timeout or cancel.
	if timeout || canceled {
		err = nil
	}

	return stdout, stderr, timeout, canceled, err
}

// An inner run container helper.
// We split these up to allow for better timeout guarantees
// (we can't fully trust Docker to timeout properly).
// This function does not try to enforce any timeouts (aside from passing along the context), that is left to callers.
// If a timeout is detected, it will be returned (but it is only one of many ways a timeout could happen).
// Returns: (stdout, stderr, timeout (only one of many types), error)
func runContainerInternal(ctx context.Context, logId log.Loggable, imageName string, mounts []MountInfo, cmd []string, baseID string) (string, string, bool, error) {
	// Get a docker client.
	// Note that cleaning this up needs to wait until after we are sure the container is dead.
	// This means we won't be defering the close right away (see cleanupRun()).
	docker, err := getDockerClient()
	if err != nil {
		return "", "", false, err
	}

	name := cleanContainerName(fmt.Sprintf("%s-%s", baseID, util.UUID()))

	dockerMounts := make([]mount.Mount, 0, len(mounts))
	for _, mount := range mounts {
		dockerMounts = append(dockerMounts, mount.ToDocker())
	}

	log.Debug("Creating container.", log.NewAttr("name", name))
	containerInstance, err := docker.ContainerCreate(
		ctx,
		&container.Config{
			Image:           imageName,
			NetworkDisabled: true,
			Cmd:             cmd,
		},
		&container.HostConfig{
			Mounts: dockerMounts,
			LogConfig: container.LogConfig{
				// Don't store any logs, we will copy stdout/stderr directly.
				Type: "none",
			},
		},
		nil,
		nil,
		name)

	if err != nil {
		docker.Close()
		return "", "", false, fmt.Errorf("Failed to create container '%s': '%w'.", name, err)
	}

	// Now that we have the container, we can schedule cleanup in the background.
	defer func() {
		go cleanupRun(docker, name, containerInstance.ID)
	}()

	// Attach to the container so we can get stdout and stderr.
	log.Trace("Attaching container.", log.NewAttr("name", name))
	connection, err := docker.ContainerAttach(ctx, containerInstance.ID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return "", "", false, fmt.Errorf("Failed to attach to container '%s' (%s): '%w'.", name, containerInstance.ID, err)
	}
	defer connection.Close()

	// Handle copying (and possibly truncating) stdout/stderr.
	outputWaitGroup := &sync.WaitGroup{}
	outputWaitGroup.Add(1)
	output := &containerOutput{}
	go handleContainerOutput(ctx, output, outputWaitGroup, connection)

	log.Trace("Starting container.", log.NewAttr("name", name))
	err = docker.ContainerStart(ctx, containerInstance.ID, container.StartOptions{})
	if err != nil {
		return "", "", false, fmt.Errorf("Failed to start container '%s' (%s): '%w'.", name, containerInstance.ID, err)
	}

	// Wait for the container to finish.
	log.Trace("Waiting for container.", log.NewAttr("name", name))
	statusChan, errorChan := docker.ContainerWait(ctx, containerInstance.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errorChan:
		if err != nil {
			// On a timeout or cancel exit this select, and try to recover the output.
			timeout := errors.Is(err, context.DeadlineExceeded)
			canceled := errors.Is(err, context.Canceled)
			if timeout || canceled {
				break
			}

			return "", "", false, fmt.Errorf("Got an error when running container '%s' (%s): '%w'.", name, containerInstance.ID, err)
		}
	case <-statusChan:
		// Waiting is complete.
	case <-ctx.Done():
		// The context finished but the result has not shown on the error chan (yet).
	}

	// Wait for output to get copied.
	log.Trace("Waiting for container output.", log.NewAttr("name", name))
	outputWaitGroup.Wait()

	log.Debug("Done with container.", log.NewAttr("name", name))

	log.Trace("Container output.",
		logId,
		log.NewAttr("container-name", name),
		log.NewAttr("container-id", containerInstance.ID),
		log.NewAttr("stdout", output.Stdout),
		log.NewAttr("stderr", output.Stderr),
		log.NewAttr("timeout", errors.Is(ctx.Err(), context.DeadlineExceeded)),
		log.NewAttr("canceled", errors.Is(ctx.Err(), context.Canceled)),
		log.NewAttr("output-truncated", output.Truncated),
		output.Err,
	)

	return output.Stdout, output.Stderr, errors.Is(ctx.Err(), context.DeadlineExceeded), nil
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
func handleContainerOutput(ctx context.Context, output *containerOutput, outputWaitGroup *sync.WaitGroup, connection types.HijackedResponse) {
	defer outputWaitGroup.Done()

	// Closing the connection should also close the reader and stop any waiting read operations.
	defer connection.Close()

	successChan := make(chan bool, 1)

	// Start trying to read in another thread.
	go func() {
		handleContainerOutputInternal(output, connection.Reader)
		successChan <- true
	}()

	// Wait for either the context or read to complete.
	select {
	case <-successChan:
		return
	case <-ctx.Done():
		return
	}
}

func handleContainerOutputInternal(output *containerOutput, containerStream io.Reader) {
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

func (this MountInfo) ToDocker() mount.Mount {
	return mount.Mount{
		Type:     "bind",
		Source:   this.Source,
		Target:   this.Target,
		ReadOnly: this.ReadOnly,
	}
}

// Set the value and return a function to reset it back to its original state.
func SetExtraInitTimeSecsForTesting(newValue int) func() {
	oldValue := extraInitTimeSecs
	extraInitTimeSecs = newValue

	return func() {
		extraInitTimeSecs = oldValue
	}
}

func killContainer(docker *client.Client, name string, id string) {
	// The container should have already gracefully exited.
	// If not, kill it without any grace.
	// Ignore any errors.
	docker.ContainerKill(context.Background(), id, "KILL")

	err := docker.ContainerRemove(context.Background(), id, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		log.Warn("Failed to remove container.", err, log.NewAttr("name", name), log.NewAttr("id", id))
	}
}

// Cleanup any leftovers from the running the container.
// This should generally be called in another go routine to prevent blocking.
func cleanupRun(docker *client.Client, containerName string, containerID string) {
	killContainer(docker, containerName, containerID)

	err := docker.Close()
	if err != nil {
		log.Warn("Failed to close docker client connection.", err)
	}
}
