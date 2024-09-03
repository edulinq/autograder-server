package grader

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const PYTHON_AUTOGRADER_INVOCATION = "python3 -m autograder.cli.grading.grade-dir --grader <grader> --dir <basedir> --outpath <outpath>"
const PYTHON_GRADER_FILENAME = "grader.py"
const PYTHON_DOCKER_IMAGE_BASENAME = "autograder.python"

func runNoDockerGrader(assignment *model.Assignment, submissionPath string, options GradeOptions, fullSubmissionID string) (
	*model.GradingInfo, map[string][]byte, string, string, error) {
	imageInfo := assignment.GetImageInfo()
	if imageInfo == nil {
		return nil, nil, "", "", fmt.Errorf("No image information associated with assignment: '%s'.", assignment.FullID())
	}

	tempDir, inputDir, outputDir, workDir, err := common.PrepTempGradingDir("nodocker")
	if err != nil {
		return nil, nil, "", "", err
	}

	if !options.LeaveTempDir {
		defer os.RemoveAll(tempDir)
	} else {
		log.Debug("Leaving behind temp grading dir.", log.NewAttr("path", tempDir))
	}

	cmd, err := getAssignmentInvocation(assignment, tempDir, inputDir, outputDir, workDir)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Copy over the static files (and do any file ops).
	err = common.CopyFileSpecs(imageInfo.BaseDir, workDir, tempDir,
		imageInfo.StaticFiles, false, imageInfo.PreStaticFileOperations, imageInfo.PostStaticFileOperations)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("Failed to copy static assignment files: '%w'.", err)
	}

	// Copy over the submission files (and do any file ops).
	err = common.CopyFileSpecs(submissionPath, inputDir, tempDir,
		[]*common.FileSpec{common.GetPathFileSpec(".")}, true, []common.FileOperation{}, imageInfo.PostSubmissionFileOperations)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("Failed to copy submission ssignment files: '%w'.", err)
	}

	stdout, stderr, err := runCMD(cmd)
	if err != nil {
		return nil, nil, stdout, stderr,
			fmt.Errorf("Failed to run non-docker grader for assignment '%s': '%w'.", assignment.FullID(), err)
	}

	resultPath := filepath.Join(outputDir, common.GRADER_OUTPUT_RESULT_FILENAME)
	if !util.PathExists(resultPath) {
		return nil, nil, stdout, stderr, fmt.Errorf("Cannot find output file ('%s') after non-docker grading.", resultPath)
	}

	var gradingInfo model.GradingInfo
	err = util.JSONFromFile(resultPath, &gradingInfo)
	if err != nil {
		return nil, nil, stdout, stderr, err
	}

	fileContents, err := util.GzipDirectoryToBytes(outputDir)
	if err != nil {
		return nil, nil, stdout, stderr, fmt.Errorf("Failed to copy grading output '%s': '%w'.", outputDir, err)
	}

	return &gradingInfo, fileContents, stdout, stderr, nil
}

func runCMD(cmd *exec.Cmd) (string, string, error) {
	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	err := cmd.Run()

	stdout := outBuffer.String()
	stderr := errBuffer.String()

	return stdout, stderr, err
}

// Get a command to invoke the non-docker grader.
func getAssignmentInvocation(assignment *model.Assignment,
	baseDir string, inputDir string, outputDir string, workDir string) (*exec.Cmd, error) {
	imageInfo := assignment.GetImageInfo()
	if imageInfo == nil {
		return nil, fmt.Errorf("No image information associated with assignment: '%s'.", assignment.FullID())
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
		return nil, fmt.Errorf("Cannot get non-docker grader invocation for assignment: '%s'.", assignment.FullID())
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

	cmd := exec.Command(cleanCommand[0], cleanCommand[1:]...)
	cmd.Dir = workDir

	return cmd, nil
}
