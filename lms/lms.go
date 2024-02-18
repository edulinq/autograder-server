package lms

import (
    "fmt"

    "github.com/edulinq/autograder/lms/backend/canvas"
    "github.com/edulinq/autograder/lms/backend/test"
    "github.com/edulinq/autograder/lms/lmstypes"
    "github.com/edulinq/autograder/model"
)

type lmsBackend interface {
    FetchAssignments() ([]*lmstypes.Assignment, error)
    FetchAssignment(assignmentID string) (*lmstypes.Assignment, error)

    UpdateComments(assignmentID string, comments []*lmstypes.SubmissionComment) error
    UpdateComment(assignmentID string, comment *lmstypes.SubmissionComment) error

    FetchAssignmentScores(assignmentID string) ([]*lmstypes.SubmissionScore, error)
    FetchAssignmentScore(assignmentID string, userID string) (*lmstypes.SubmissionScore, error)
    UpdateAssignmentScores(assignmentID string, scores []*lmstypes.SubmissionScore) error

    FetchUsers() ([]*lmstypes.User, error)
    FetchUser(email string) (*lmstypes.User, error)
}

func getBackend(course *model.Course) (lmsBackend, error) {
    adapter := course.GetLMSAdapter();
    if (adapter == nil) {
        return nil, fmt.Errorf("Course '%s' has no LMS information.", course.GetID());
    }

    switch (adapter.Type) {
        case model.LMS_TYPE_CANVAS:
            backend, err := canvas.NewBackend(adapter.LMSCourseID, adapter.APIToken, adapter.BaseURL);
            if (err != nil) {
                return nil, err;
            }

            return backend, nil;
        case model.LMS_TYPE_TEST:
            backend, err := test.NewBackend(course.GetID());
            if (err != nil) {
                return nil, err;
            }

            return backend, nil;
        default:
            return nil, fmt.Errorf("Unknown LMS type: '%s'.", adapter.Type);
    }
}

func FetchAssignment(course *model.Course, assignmentID string) (*lmstypes.Assignment, error) {
    backend, err := getBackend(course);
    if (err != nil) {
        return nil, err;
    }

    return backend.FetchAssignment(assignmentID);
}

func FetchAssignments(course *model.Course) ([]*lmstypes.Assignment, error) {
    backend, err := getBackend(course);
    if (err != nil) {
        return nil, err;
    }

    return backend.FetchAssignments();
}

func UpdateComments(course *model.Course, assignmentID string, comments []*lmstypes.SubmissionComment) error {
    backend, err := getBackend(course);
    if (err != nil) {
        return err;
    }

    return backend.UpdateComments(assignmentID, comments);
}

func UpdateComment(course *model.Course, assignmentID string, comment *lmstypes.SubmissionComment) error {
    backend, err := getBackend(course);
    if (err != nil) {
        return err;
    }

    return backend.UpdateComment(assignmentID, comment);
}

func FetchAssignmentScores(course *model.Course, assignmentID string) ([]*lmstypes.SubmissionScore, error) {
    backend, err := getBackend(course);
    if (err != nil) {
        return nil, err;
    }

    return backend.FetchAssignmentScores(assignmentID);
}

func FetchAssignmentScore(course *model.Course, assignmentID string, userID string) (*lmstypes.SubmissionScore, error) {
    backend, err := getBackend(course);
    if (err != nil) {
        return nil, err;
    }

    return backend.FetchAssignmentScore(assignmentID, userID);
}

func UpdateAssignmentScores(course *model.Course, assignmentID string, scores []*lmstypes.SubmissionScore) error {
    backend, err := getBackend(course);
    if (err != nil) {
        return err;
    }

    return backend.UpdateAssignmentScores(assignmentID, scores);
}

func FetchUsers(course *model.Course, ) ([]*lmstypes.User, error) {
    backend, err := getBackend(course);
    if (err != nil) {
        return nil, err;
    }

    return backend.FetchUsers();
}

func FetchUser(course *model.Course, email string) (*lmstypes.User, error) {
    backend, err := getBackend(course);
    if (err != nil) {
        return nil, err;
    }

    return backend.FetchUser(email);
}

