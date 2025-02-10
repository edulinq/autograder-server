package grader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const PYTHON_AUTOGRADER_INVOCATION = "python3 -m autograder.cli.grading.grade-dir --grader <grader> --dir <basedir> --outpath <outpath>"
const PYTHON_GRADER_FILENAME = "grader.py"
const PYTHON_DOCKER_IMAGE_BASENAME = "edulinq/grader.python"

// A small delay to wait for a process to finish after already timing out.
var noDockerTimeoutWaitDelayMS int = 10 * 1000

// Returns: (result, file contents, stdout, stderr, failure message (soft failure), error (hard failure)).
func runNoDockerGrader(ctx context.Context, assignment *model.Assignment, submissionPath string, options GradeOptions, fullSubmissionID string) (
	*model.GradingInfo, map[string][]byte, string, string, string, error) {
	imageInfo := assignment.GetImageInfo()
	if imageInfo == nil {
		return nil, nil, "", "", "", fmt.Errorf("No image information associated with assignment: '%s'.", assignment.FullID())
	}

	tempDir, inputDir, outputDir, workDir, err := common.PrepTempGradingDir("nodocker")
	if err != nil {
		return nil, nil, "", "", "", err
	}

	if !options.LeaveTempDir {
		defer util.RemoveDirent(tempDir)
	} else {
		log.Debug("Leaving behind temp grading dir.", log.NewAttr("path", tempDir))
	}

	ctx, cmd, err := getAssignmentInvocation(ctx, assignment, tempDir, inputDir, outputDir, workDir)
	if err != nil {
		return nil, nil, "", "", "", err
	}

	// Copy over the static files (and do any file ops).
	err = common.CopyFileSpecsWithOps(imageInfo.BaseDirFunc(), workDir, tempDir,
		imageInfo.StaticFiles, false, imageInfo.PreStaticFileOperations, imageInfo.PostStaticFileOperations)
	if err != nil {
		return nil, nil, "", "", "", fmt.Errorf("Failed to copy static assignment files: '%w'.", err)
	}

	// Copy over the submission files (and do any file ops).
	err = common.CopyFileSpecsWithOps(submissionPath, inputDir, tempDir,
		[]*common.FileSpec{common.GetPathFileSpec(".")}, true, []*common.FileOperation{}, imageInfo.PostSubmissionFileOperations)
	if err != nil {
		return nil, nil, "", "", "", fmt.Errorf("Failed to copy submission ssignment files: '%w'.", err)
	}

	stdout, stderr, timeout, canceled, err := runCMD(ctx, cmd)
	if err != nil {
		return nil, nil, stdout, stderr, "",
			fmt.Errorf("Failed to run non-docker grader for assignment '%s': '%w'.", assignment.FullID(), err)
	}

	if timeout {
		return nil, nil, stdout, stderr, getTimeoutMessage(assignment), nil
	}

	if canceled {
		return nil, nil, stdout, stderr, getCanceledMessage(assignment), nil
	}

	resultPath := filepath.Join(outputDir, common.GRADER_OUTPUT_RESULT_FILENAME)
	if !util.PathExists(resultPath) {
		return nil, nil, stdout, stderr, "", fmt.Errorf("Cannot find output file ('%s') after non-docker grading.", resultPath)
	}

	var gradingInfo model.GradingInfo
	err = util.JSONFromFile(resultPath, &gradingInfo)
	if err != nil {
		return nil, nil, stdout, stderr, "", err
	}

	fileContents, err := util.GzipDirectoryToBytes(outputDir)
	if err != nil {
		return nil, nil, stdout, stderr, "", fmt.Errorf("Failed to copy grading output '%s': '%w'.", outputDir, err)
	}

	return &gradingInfo, fileContents, stdout, stderr, "", nil
}

func runCMD(ctx context.Context, cmd *exec.Cmd) (string, string, bool, bool, error) {
	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	timeout := false
	canceled := false

	err := cmd.Run()
	if err != nil {
		timeout = errors.Is(ctx.Err(), context.DeadlineExceeded)
		canceled = errors.Is(ctx.Err(), context.Canceled)
	}

	if timeout || canceled {
		err = nil
	}

	stdout := outBuffer.String()
	stderr := errBuffer.String()

	return stdout, stderr, timeout, canceled, err
}

// Get a command to invoke the non-docker grader.
func getAssignmentInvocation(ctx context.Context, assignment *model.Assignment,
	baseDir string, inputDir string, outputDir string, workDir string) (context.Context, *exec.Cmd, error) {
	imageInfo := assignment.GetImageInfo()
	if imageInfo == nil {
		return ctx, nil, fmt.Errorf("No image information associated with assignment: '%s'.", assignment.FullID())
	}

	var rawCommand []string = nil

	if (imageInfo.Invocation != nil) && (len(imageInfo.Invocation) > 0) {
		rawCommand = imageInfo.Invocation
	}

	// Special case for the python grader (we know how to invoke that).
	if (rawCommand == nil) && (strings.Contains(imageInfo.Image, PYTHON_DOCKER_IMAGE_BASENAME)) {
		rawCommand = strings.Split(PYTHON_AUTOGRADER_INVOCATION, " ")
	}

	if rawCommand == nil {
		return ctx, nil, fmt.Errorf("Cannot get non-docker grader invocation for assignment: '%s'.", assignment.FullID())
	}

	cleanCommand := make([]string, 0, len(rawCommand))
	for _, value := range rawCommand {
		if value == "<grader>" {
			value = filepath.Join(workDir, PYTHON_GRADER_FILENAME)
		} else if value == "<basedir>" {
			value = baseDir
		} else if value == "<inputdir>" {
			value = inputDir
		} else if value == "<outputdir>" {
			value = outputDir
		} else if value == "<workdir>" {
			value = workDir
		} else if value == "<outpath>" {
			value = filepath.Join(outputDir, common.GRADER_OUTPUT_RESULT_FILENAME)
		}

		cleanCommand = append(cleanCommand, value)
	}

	// Set a timeout for the command using the existing context as the parent.
	var cancelFunc context.CancelFunc = nil
	if assignment.MaxRuntimeSecs > 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, time.Duration(assignment.MaxRuntimeSecs)*time.Second)
	}

	cmd := exec.CommandContext(ctx, cleanCommand[0], cleanCommand[1:]...)
	cmd.Dir = workDir

	// Ensure the timeout context is canceled.
	if cancelFunc != nil {
		oldCancel := cmd.Cancel
		cmd.Cancel = func() error {
			cmd.WaitDelay = time.Duration(noDockerTimeoutWaitDelayMS) * time.Millisecond
			defer cancelFunc()
			return oldCancel()
		}
	}

	return ctx, cmd, nil
}

// Set the value and return a function to reset it back to its original state.
func SetNoDockerTimeoutWaitDelayMSForTesting(newValue int) func() {
	oldValue := noDockerTimeoutWaitDelayMS
	noDockerTimeoutWaitDelayMS = newValue

	return func() {
		noDockerTimeoutWaitDelayMS = oldValue
	}
}
