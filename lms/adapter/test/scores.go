package test

import (
    "fmt"

    "github.com/eriq-augustine/autograder/lms"
)

func (this *TestLMSAdapter) UpdateAssignmentScores(assignmentID string, scores []*lms.SubmissionScore) error {
    if (this.FailUpdateAssignmentScores) {
        return fmt.Errorf("Induced Failure");
    }

    return nil;
}
