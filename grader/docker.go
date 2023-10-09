package grader

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

func BuildDockerImagesJoinErrors(buildOptions *docker.BuildOptions) ([]string, error) {
    imageNames, errs := BuildDockerImages(buildOptions);
    return imageNames, errors.Join(errs...);
}

func BuildDockerImages(buildOptions *docker.BuildOptions) ([]string, []error) {
    errs := make([]error, 0);
    imageNames := make([]string, 0);

    for _, course := range courses {
        for _, assignment := range course.Assignments {
            err := docker.BuildImageWithOptions(assignment.GetImageInfo(), buildOptions);
            if (err != nil) {
                errs = append(errs, fmt.Errorf("Failed to build docker grader image for assignment (%s): '%w'.", assignment.FullID(), err));
            } else {
                imageNames = append(imageNames, assignment.ImageName());
            }
        }
    }

    return imageNames, errs;
}

// Grade using a docker container.
// Directory information:
//  - input -- A temp dir that will be mounted at DOCKER_INPUT_DIR (read-only).
//  - output -- Passed in directory that will be mounted at DOCKER_OUTPUT_DIR.
//  - work -- Should already be created inside the docker image, will only exist within the container.
func RunDockerGrader(assignment *model.Assignment, submissionPath string, outputDir string, options GradeOptions, gradingID string) (*model.GradedAssignment, string, error) {
    os.MkdirAll(outputDir, 0755);
    if (!util.IsEmptyDir(outputDir)) {
        return nil, "", fmt.Errorf("Output dir for docker grader is not empty.");
    }

    // Create a temp directory to use for input (will be mounted to the container).
    tempInputDir, err := os.MkdirTemp("", "autograding-docker-input-");
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

    output, err := docker.RunContainer(assignment.ImageName(), tempInputDir, outputDir, gradingID);
    if (err != nil) {
        return nil, "", err;
    }

    resultPath := filepath.Join(outputDir, common.GRADER_OUTPUT_RESULT_FILENAME);
    if (!util.PathExists(resultPath)) {
        return nil, output, fmt.Errorf("Cannot find output file ('%s') after the grading container (%s) was run.", resultPath, assignment.ImageName());
    }

    var result model.GradedAssignment;
    err = util.JSONFromFile(resultPath, &result);
    if (err != nil) {
        return nil, output, err;
    }

    return &result, output, nil;
}
