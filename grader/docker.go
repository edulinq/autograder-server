package grader

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// Grade using a docker container.
// Directory information:
//  - input -- A temp dir that will be mounted at DOCKER_INPUT_DIR (read-only).
//  - output -- Passed in directory that will be mounted at DOCKER_OUTPUT_DIR.
//  - work -- Should already be created inside the docker image, will only exist within the container.
func runDockerGrader(assignment *model.Assignment, submissionPath string, options GradeOptions, fullSubmissionID string) (
        *model.GradingInfo, map[string][]byte, string, string, error) {
    tempDir, inputDir, outputDir, _, err := common.PrepTempGradingDir("docker");
    if (err != nil) {
        return nil, nil, "", "", err;
    }

    if (!options.LeaveTempDir) {
        defer os.RemoveAll(tempDir);
    } else {
        log.Info("Leaving behind temp grading dir.", log.NewAttr("path", tempDir));
    }

    // Copy over submission files to the temp input dir.
    err = util.CopyDirent(submissionPath, inputDir, true);
    if (err != nil) {
        return nil, nil, "", "", fmt.Errorf("Failed to copy over submission/input contents: '%w'.", err);
    }

    stdout, stderr, err := docker.RunContainer(assignment.ImageName(), inputDir, outputDir, fullSubmissionID);
    if (err != nil) {
        return nil, nil, stdout, stderr, err;
    }

    resultPath := filepath.Join(outputDir, common.GRADER_OUTPUT_RESULT_FILENAME);
    if (!util.PathExists(resultPath)) {
        return nil, nil, stdout, stderr,
                fmt.Errorf("Cannot find output file ('%s') after the grading container (%s) was run.", resultPath, assignment.ImageName());
    }

    var gradingInfo model.GradingInfo;
    err = util.JSONFromFile(resultPath, &gradingInfo);
    if (err != nil) {
        return nil, nil, stdout, stderr, err;
    }

    fileContents, err := util.GzipDirectoryToBytes(outputDir);
    if (err != nil) {
        return nil, nil, stdout, stderr, fmt.Errorf("Failed to copy grading output '%s': '%w'.", outputDir, err);
    }

    return &gradingInfo, fileContents, stdout, stderr, nil;
}
