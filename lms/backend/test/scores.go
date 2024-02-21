package test

import (
    "fmt"

    "github.com/edulinq/autograder/lms/lmstypes"
)

func (this *TestLMSBackend) UpdateAssignmentScores(assignmentID string, scores []*lmstypes.SubmissionScore) error {
    if (failUpdateAssignmentScores) {
        return fmt.Errorf("Induced Failure");
    }

    return nil;
}
