package grader

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

func BuildDockerImages(force bool, buildOptions *docker.BuildOptions) ([]string, map[string]error) {
    goodImageNames := make([]string, 0);
    errors := make(map[string]error);

    for _, course := range courses {
        courseGoodImageNames, courseErrors := course.BuildAssignmentImages(force, false, buildOptions);

        goodImageNames = append(goodImageNames, courseGoodImageNames...);

        for key, value := range courseErrors {
            errors[key] = value;
        }
    }

    return goodImageNames, errors;
}

// Grade using a docker container.
// Directory information:
//  - input -- A temp dir that will be mounted at DOCKER_INPUT_DIR (read-only).
//  - output -- Passed in directory that will be mounted at DOCKER_OUTPUT_DIR.
//  - work -- Should already be created inside the docker image, will only exist within the container.
func RunDockerGrader(assignment model.Assignment, submissionPath string, outputDir string, options GradeOptions, fullSubmissionID string) (*artifact.GradedAssignment, string, error) {
    os.MkdirAll(outputDir, 0755);
    if (!util.IsEmptyDir(outputDir)) {
        return nil, "", fmt.Errorf("Output dir for docker grader is not empty.");
    }

    // Create a temp directory to use for input (will be mounted to the container).
    tempInputDir, err := util.MkDirTemp("autograding-docker-input-");
    if (err != nil) {
        return nil, "", fmt.Errorf("Could not create temp input dir: '%w'.", err);
    }

    if (options.LeaveTempDir) {
        log.Info().Str("path", tempInputDir).Msg("Leaving behind temp input dir.");
    } else {
        defer os.RemoveAll(tempInputDir);
    }

    // Copy over submission files to the temp input dir.
    err = util.CopyDirent(submissionPath, tempInputDir, true);
    if (err != nil) {
        return nil, "", fmt.Errorf("Failed to copy over submission/input contents: '%w'.", err);
    }

    output, err := docker.RunContainer(assignment.ImageName(), tempInputDir, outputDir, fullSubmissionID);
    if (err != nil) {
        return nil, "", err;
    }

    resultPath := filepath.Join(outputDir, common.GRADER_OUTPUT_RESULT_FILENAME);
    if (!util.PathExists(resultPath)) {
        return nil, output, fmt.Errorf("Cannot find output file ('%s') after the grading container (%s) was run.", resultPath, assignment.ImageName());
    }

    var result artifact.GradedAssignment;
    err = util.JSONFromFile(resultPath, &result);
    if (err != nil) {
        return nil, output, err;
    }

    return &result, output, nil;
}
