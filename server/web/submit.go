package web

import (
    "fmt"
    "net/http"
    "os"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

type SubmissionRequest struct {
    model.BaseAPIRequest
    Assignment string `json:"assignment"`

    Dir string `json:"-"`
}

func (this *SubmissionRequest) String() string {
    return util.BaseString(this);
}

func NewSubmissionRequest(request *http.Request) (*SubmissionRequest, *model.APIResponse, error) {
    var apiRequest SubmissionRequest;

    err := model.APIRequestFromHTTP(&apiRequest, request);
    if (err != nil) {
        return nil, nil, err;
    }

    ok, err := grader.AuthAPIRequest(&apiRequest.BaseAPIRequest);
    if (err != nil) {
        return nil, nil, err;
    } else if (!ok) {
        return nil, model.NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    apiRequest.Dir, err = model.StoreRequestFiles(request);
    if (err != nil) {
        return nil, nil, err;
    }

    return &apiRequest, nil, nil;
}

func (this *SubmissionRequest) Close() error {
    return os.RemoveAll(this.Dir);
}

func (this *SubmissionRequest) Clean() error {
    var err error;
    this.Assignment, err = model.ValidateID(this.Assignment);
    if (err != nil) {
        return fmt.Errorf("Could not clean SubmissionRequest assignment ID ('%s'): '%w'.", this.Assignment, err);
    }

    return nil;
}

func handleSubmit(submission *SubmissionRequest) (int, any, error) {
    assignment := grader.GetAssignment(submission.Course, submission.Assignment);
    if (assignment == nil) {
        return http.StatusBadRequest, fmt.Sprintf("Could not find assignment ('%s') for course ('%s').", submission.Assignment, submission.Course,), nil;
    }

    result, err := grader.GradeDefault(assignment, submission.Dir, submission.User);
    if (err != nil) {
        return 0, nil, err;
    }

    return 0, result, nil;
}
