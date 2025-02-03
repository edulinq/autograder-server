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
	"time"

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
	// Set a timeout for the container.
	// We are doing this before actually starting the container to (hopefully) stop any long running when creating or starting the container.
	// Note that this timeout context is not based on the passed-in context.
	// We will add a small amount to the timeout to compensate for using the same timeout when creating and starting.
	var timeoutCtx context.Context = context.Background()
	var cancel context.CancelFunc
	successChan := make(chan bool, 1)

	if maxRuntimeSecs > 0 {
		timeoutCtx, cancel = context.WithTimeout(context.Background(), time.Duration(maxRuntimeSecs+extraInitTimeSecs)*time.Second)
		defer cancel()
	}

	var stdout string
	var stderr string
	var timeout bool
	var canceled bool
	var err error

	go func() {
		stdout, stderr, timeout, canceled, err = runContainerInternal(ctx, logId, imageName, mounts, cmd, baseID, maxRuntimeSecs)
		successChan <- true
	}()

	select {
	case <-successChan:
		// Success
		return stdout, stderr, timeout, canceled, err
	case <-timeoutCtx.Done():
		// Timeout
		return "", "", true, false, nil
	}

	return "", "", false, false, fmt.Errorf("Docker container run (and timeout) failed.")
}

// An inner run container helper.
// We split these up to allow for better timeout guarantees
// (we can't fully trust Docker to timeout properly).
func runContainerInternal(ctx context.Context, logId log.Loggable, imageName string, mounts []MountInfo, cmd []string, baseID string, maxRuntimeSecs int) (string, string, bool, bool, error) {
	docker, err := getDockerClient()
	if err != nil {
		return "", "", false, false, err
	}
	defer docker.Close()

	name := cleanContainerName(fmt.Sprintf("%s-%s", baseID, util.UUID()))

	timeout := false
	canceled := false

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
		timeout = errors.Is(err, context.DeadlineExceeded)
		canceled = errors.Is(err, context.Canceled)
		if timeout || canceled {
			return "", "", timeout, canceled, nil
		}

		return "", "", false, false, fmt.Errorf("Failed to create container '%s': '%w'.", name, err)
	}

	// Ensure the container is removed (in the background).
	defer func() {
		go killContainer(docker, name, containerInstance.ID)
	}()

	// Attach to the container so we can get stdout and stderr.
	log.Trace("Attaching container.", log.NewAttr("name", name))
	connection, err := docker.ContainerAttach(ctx, containerInstance.ID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		timeout = errors.Is(err, context.DeadlineExceeded)
		canceled = errors.Is(err, context.Canceled)
		if timeout || canceled {
			return "", "", timeout, canceled, nil
		}

		return "", "", false, false, fmt.Errorf("Failed to attach to container '%s' (%s): '%w'.", name, containerInstance.ID, err)
	}
	defer connection.Conn.Close()

	// Handle copying (and possibly truncating) stdout/stderr.
	outputWaitGroup := &sync.WaitGroup{}
	outputWaitGroup.Add(1)
	output := &containerOutput{}
	go handleContainerOutput(output, outputWaitGroup, connection.Reader)

	// Set a timeout for the container.
	var cancel context.CancelFunc
	if maxRuntimeSecs > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(maxRuntimeSecs)*time.Second)
		defer cancel()
	}

	log.Trace("Starting container.", log.NewAttr("name", name))
	err = docker.ContainerStart(ctx, containerInstance.ID, container.StartOptions{})
	if err != nil {
		timeout = errors.Is(err, context.DeadlineExceeded)
		canceled = errors.Is(err, context.Canceled)
		if timeout || canceled {
			return "", "", timeout, canceled, nil
		}

		return "", "", false, false, fmt.Errorf("Failed to start container '%s' (%s): '%w'.", name, containerInstance.ID, err)
	}

	// Wait for the container to finish.
	log.Trace("Waiting for container.", log.NewAttr("name", name))
	statusChan, errorChan := docker.ContainerWait(ctx, containerInstance.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errorChan:
		if err != nil {
			// On a timeout or cancel exit this select, and try to recover the output.
			timeout = errors.Is(err, context.DeadlineExceeded)
			canceled = errors.Is(err, context.Canceled)
			if timeout || canceled {
				break
			}

			return "", "", false, false, fmt.Errorf("Got an error when running container '%s' (%s): '%w'.", name, containerInstance.ID, err)
		}
	case <-statusChan:
		// Waiting is complete.
	case <-ctx.Done():
		// The context finished but the result has not shown on the error chan (yet).
	}

	timeout = timeout || errors.Is(ctx.Err(), context.DeadlineExceeded)
	canceled = canceled || errors.Is(ctx.Err(), context.Canceled)

	// If the container has been canceled, then don't even try to recover the output.
	// We do not want to wait on reading from the output stream.
	if canceled {
		return output.Stdout, output.Stderr, timeout, canceled, nil
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
		log.NewAttr("timeout", timeout),
		log.NewAttr("canceled", canceled),
		log.NewAttr("output-truncated", output.Truncated),
		output.Err,
	)

	return output.Stdout, output.Stderr, timeout, canceled, nil
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
	log.Trace("killing container.", log.NewAttr("name", name), log.NewAttr("id", id))
	docker.ContainerKill(context.Background(), id, "KILL")

	err := docker.ContainerRemove(context.Background(), id, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		log.Warn("Failed to remove container.", err, log.NewAttr("name", name), log.NewAttr("id", id))
	}
}
