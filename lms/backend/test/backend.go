// A fake LMS backend for testing that reads config from a test course directory.
package test

import (
    "fmt"

    "github.com/eriq-augustine/autograder/lms/lmstypes"
)

// Settings to help in testing.
var failUpdateAssignmentScores bool = false;
var usersModifier FetchUsersModifier = nil;

type TestLMSBackend struct {
    CourseID string
}

// Change the users returned from FetchUsers() for testing.
type FetchUsersModifier func ([]*lmstypes.User) []*lmstypes.User;

func NewBackend(courseID string) (*TestLMSBackend, error) {
    if (courseID == "") {
        return nil, fmt.Errorf("Test LMS backend must have a non-empty course id.");
    }

    backend := TestLMSBackend{
        CourseID: courseID,
    };

    return &backend, nil;
}

func SetFailUpdateAssignmentScores(value bool) {
    failUpdateAssignmentScores = value;
}

func SetUsersModifier(modifier FetchUsersModifier) {
    usersModifier = modifier;
}

func ClearUsersModifier() {
    usersModifier = nil;
}

func (this *TestLMSBackend) FetchAssignments() ([]*lmstypes.Assignment, error) {
    return nil, nil;
}

func (this *TestLMSBackend) FetchAssignment(assignmentID string) (*lmstypes.Assignment, error) {
    return nil, nil;
}

func (this *TestLMSBackend) UpdateComments(assignmentID string, comments []*lmstypes.SubmissionComment) error {
    return nil;
}

func (this *TestLMSBackend) UpdateComment(assignmentID string, comment *lmstypes.SubmissionComment) error {
    return nil;
}

func (this *TestLMSBackend) FetchAssignmentScores(assignmentID string) ([]*lmstypes.SubmissionScore, error) {
    return nil, nil;
}

func (this *TestLMSBackend) FetchAssignmentScore(assignmentID string, userID string) (*lmstypes.SubmissionScore, error) {
    return nil, nil;
}
