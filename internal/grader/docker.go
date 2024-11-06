package grader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// Grade using a docker container.
// Directory information:
//   - input -- A temp dir that will be mounted at DOCKER_INPUT_DIR (read-only).
//   - output -- Passed in directory that will be mounted at DOCKER_OUTPUT_DIR.
//   - work -- Should already be created inside the docker image, will only exist within the container.
//
// Returns: (result, file contents, stdout, stderr, failure message (soft failure), error (hard failure)).
func runDockerGrader(assignment *model.Assignment, submissionPath string, options GradeOptions, fullSubmissionID string) (
	*model.GradingInfo, map[string][]byte, string, string, string, error) {
	tempDir, inputDir, outputDir, _, err := common.PrepTempGradingDir("docker")
	if err != nil {
		return nil, nil, "", "", "", err
	}

	if !options.LeaveTempDir {
		defer os.RemoveAll(tempDir)
	} else {
		log.Debug("Leaving behind temp grading dir.", assignment, log.NewAttr("path", tempDir))
	}

	// Copy over submission files to the temp input dir.
	err = util.CopyDirent(submissionPath, inputDir, true)
	if err != nil {
		return nil, nil, "", "", "", fmt.Errorf("Failed to copy over submission/input contents: '%w'.", err)
	}

	stdout, stderr, timeout, err := docker.RunContainer(assignment, assignment.ImageName(), inputDir, outputDir, fullSubmissionID, assignment.MaxRuntimeSecs)
	if err != nil {
		return nil, nil, stdout, stderr, "", err
	}

	if timeout {
		return nil, nil, stdout, stderr, fmt.Sprintf("Submission has ran for too long and was killed. Max assignment runtime is %d seconds (server hard limit is %d seconds). Check for infinite loops/recursion and consult with your instructors/TAs.", assignment.MaxRuntimeSecs, config.DOCKER_RUNTIME_MAX_SECS.Get()), nil
	}

	resultPath := filepath.Join(outputDir, common.GRADER_OUTPUT_RESULT_FILENAME)
	if !util.PathExists(resultPath) {
		return nil, nil, stdout, stderr, "",
			fmt.Errorf("Cannot find output file ('%s') after the grading container (%s) was run.", resultPath, assignment.ImageName())
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
