package grader

// Handle running docker containers for grading.

import (
    "fmt"

    "github.com/eriq-augustine/autograder/model"
)

// TEST
func RunDockerGrader(assignment *model.Assignment, submissionPath string, outputDir string, options GradeOptions) (*model.GradedAssignment, error) {
    // TEST
    return nil, fmt.Errorf("No implemented");
}
