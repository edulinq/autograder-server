package test

// A fake LMS adapter for testing that reads config from a test course directory.

import (
    "fmt"

    "github.com/eriq-augustine/autograder/lms"
)

type TestLMSAdapter struct {
    SourceCourse lms.SourceCourse
}

func NewAdapter(sourceCourse lms.SourceCourse) (*TestLMSAdapter, error) {
    if (sourceCourse == nil) {
        return nil, fmt.Errorf("Test adapter must have a non-nil source course.");
    }

    adapter := TestLMSAdapter{
        SourceCourse: sourceCourse,
    };

    return &adapter, nil;
}

func (this *TestLMSAdapter) FetchAssignment(assignmentID string) (*lms.Assignment, error) {
    return nil, nil;
}

func (this *TestLMSAdapter) UpdateComments(assignmentID string, comments []*lms.SubmissionComment) error {
    return nil;
}

func (this *TestLMSAdapter) UpdateComment(assignmentID string, comment *lms.SubmissionComment) error {
    return nil;
}

func (this *TestLMSAdapter) FetchAssignmentScores(assignmentID string) ([]*lms.SubmissionScore, error) {
    return nil, nil;
}

func (this *TestLMSAdapter) UpdateAssignmentScores(assignmentID string, scores []*lms.SubmissionScore) error {
    return nil;
}
