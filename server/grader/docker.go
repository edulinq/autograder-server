package grader

import (
    "errors"
    "fmt"

    "github.com/eriq-augustine/autograder/model"
)

func BuildDockerImagesJoinErrors(buildOptions *model.DockerBuildOptions) error {
    return errors.Join(BuildDockerImages(buildOptions)...);
}

func BuildDockerImages(buildOptions *model.DockerBuildOptions) []error {
    errs := make([]error, 0);

    for _, course := range courses {
        for _, assignment := range course.Assignments {
            err := assignment.BuildDockerImageWithOptions(buildOptions);
            if (err != nil) {
                errs = append(errs, fmt.Errorf("Failed to build docker grader image for course (%s) assignment (%s): '%w'.", assignment.FullID(), course.ID, err));
            }
        }
    }

    return errs;
}

// TEST
func RunDockerGrader(assignment *model.Assignment, submissionPath string, outputDir string, options GradeOptions) (*model.GradedAssignment, error) {
    // TEST
    return nil, fmt.Errorf("No implemented");
}
