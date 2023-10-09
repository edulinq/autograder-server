package web

import (
    "fmt"
    "net/http"
    "os"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

type SubmissionRequest struct {
    BaseAPIRequest
    Assignment string `json:"assignment"`
    Message string `json:"message"`

    Dir string `json:"-"`
}

type SubmissionResponse struct {
    Summary *model.SubmissionSummary `json:"summary"`
    Assignment *model.GradedAssignment `json:"result"`
}

func (this *SubmissionRequest) String() string {
    return util.BaseString(this);
}

func NewSubmissionRequest(request *http.Request) (*SubmissionRequest, *APIResponse, error) {
    var apiRequest SubmissionRequest;
    err := APIRequestFromPOST(&apiRequest, request);
    if (err != nil) {
        return nil, nil, err;
    }

    course, _, err := grader.VerifyCourseAssignment(apiRequest.Course, apiRequest.Assignment);
    if (err != nil) {
        return nil, nil, err;
    }

    ok, _, err := AuthAPIRequest(&apiRequest.BaseAPIRequest, course);
    if (err != nil) {
        return nil, nil, err;
    } else if (!ok) {
        return nil, NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    apiRequest.Dir, err = StoreRequestFiles(request);
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
    this.Assignment, err = common.ValidateID(this.Assignment);
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

    result, summary, output, err := grader.GradeDefault(assignment, submission.Dir, submission.User, submission.Message);
    if (err != nil) {
        if (output != "") {
            log.Debug().Err(err).Str("output", output).Msg("Submission grading failed, but output exists.");
            return 0, NewSoftFailureResponse(output), nil;
        }

        return 0, nil, err;
    }

    response := SubmissionResponse{
        Summary: summary,
        Assignment: result,
    };

    return 0, response, nil;
}
