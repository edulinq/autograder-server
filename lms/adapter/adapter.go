package adapter

import (
    "fmt"
    "strings"

    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/lms/adapter/canvas"
    "github.com/eriq-augustine/autograder/lms/adapter/test"
)

const (
    LMS_TYPE_CANVAS = "canvas"
    LMS_TYPE_TEST = "test"
)

type LMSAdapter struct {
    Type string `json:"type"`

    // Connection options.
    CourseID string `json:"course-id"`
    APIToken string `json:"api-token"`
    BaseURL string `json:"base-url"`

    // Behavior options.
    SyncUserAttributes bool `json:"sync-user-attributes"`
    SyncAddUsers bool `json:"sync-add-users"`
    SyncRemoveUsers bool `json:"sync-remove-users"`

    adapter Backend `json:"-"`
}

type Backend interface {
    FetchAssignment(assignmentID string) (*lms.Assignment, error)

    UpdateComments(assignmentID string, comments []*lms.SubmissionComment) error
    UpdateComment(assignmentID string, comment *lms.SubmissionComment) error

    FetchAssignmentScores(assignmentID string) ([]*lms.SubmissionScore, error)
    UpdateAssignmentScores(assignmentID string, scores []*lms.SubmissionScore) error

    FetchUsers() ([]*lms.User, error)
    FetchUser(email string) (*lms.User, error)
}

func (this *LMSAdapter) Validate(source lms.SourceCourse) error {
    if (this.Type == "") {
        return fmt.Errorf("LMS type cannot be empty.");
    }
    this.Type = strings.ToLower(this.Type);

    if (this.Type == LMS_TYPE_CANVAS) {
        adapter, err := canvas.NewAdapter(this.CourseID, this.APIToken, this.BaseURL);
        if (err != nil) {
            return err;
        }

        this.adapter = adapter;
    } else if (this.Type == LMS_TYPE_TEST) {
        adapter, err := test.NewAdapter(source);
        if (err != nil) {
            return err;
        }

        this.adapter = adapter;
    } else {
        return fmt.Errorf("Unknown LMS type: '%s'.", this.Type);
    }

    return nil;
}

func (this *LMSAdapter) FetchAssignment(assignmentID string) (*lms.Assignment, error) {
    return this.adapter.FetchAssignment(assignmentID);
}

func (this *LMSAdapter) UpdateComments(assignmentID string, comments []*lms.SubmissionComment) error {
    return this.adapter.UpdateComments(assignmentID, comments);
}

func (this *LMSAdapter) UpdateComment(assignmentID string, comment *lms.SubmissionComment) error {
    return this.adapter.UpdateComment(assignmentID, comment);
}

func (this *LMSAdapter) FetchAssignmentScores(assignmentID string) ([]*lms.SubmissionScore, error) {
    return this.adapter.FetchAssignmentScores(assignmentID);
}

func (this *LMSAdapter) UpdateAssignmentScores(assignmentID string, scores []*lms.SubmissionScore) error {
    return this.adapter.UpdateAssignmentScores(assignmentID, scores);
}

func (this *LMSAdapter) FetchUsers() ([]*lms.User, error) {
    return this.adapter.FetchUsers();
}

func (this *LMSAdapter) FetchUser(email string) (*lms.User, error) {
    return this.adapter.FetchUser(email);
}

