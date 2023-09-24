package web

import (
    "fmt"
    "net/http"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

type PeekRequest struct {
    model.BaseAPIRequest
    Assignment string `json:"assignment"`
}

type PeekResponse struct {
    Assignment *model.GradedAssignment `json:"result"`
}

func (this *PeekRequest) String() string {
    return util.BaseString(this);
}

func NewPeekRequest(request *http.Request) (*PeekRequest, *model.APIResponse, error) {
    var apiRequest PeekRequest;
    err := model.APIRequestFromPOST(&apiRequest, request);
    if (err != nil) {
        return nil, nil, err;
    }

    course, _, err := grader.VerifyCourseAssignment(apiRequest.Course, apiRequest.Assignment);
    if (err != nil) {
        return nil, nil, err;
    }

    ok, err := AuthAPIRequest(&apiRequest.BaseAPIRequest, course);
    if (err != nil) {
        return nil, nil, err;
    } else if (!ok) {
        return nil, model.NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    return &apiRequest, nil, nil;
}

func (this *PeekRequest) Close() error {
    return nil;
}

func (this *PeekRequest) Clean() error {
    var err error;
    this.Assignment, err = model.ValidateID(this.Assignment);
    if (err != nil) {
        return fmt.Errorf("Could not clean PeekRequest assignment ID ('%s'): '%w'.", this.Assignment, err);
    }

    return nil;
}

func handlePeek(request *PeekRequest) (int, any, error) {
    assignment := grader.GetAssignment(request.Course, request.Assignment);
    if (assignment == nil) {
        return http.StatusBadRequest, fmt.Sprintf("Could not find assignment ('%s') for course ('%s').", request.Assignment, request.Course,), nil;
    }

    paths, err := assignment.GetSubmissionResults(request.User);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed tp get submission results: '%w'.", err);
    }

    response := PeekResponse{};

    if (len(paths) == 0) {
        return 0, response, nil;
    }

    path := paths[len(paths) - 1]

    result := model.GradedAssignment{};
    err = util.JSONFromFile(path, &result);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to deserialize result '%s': '%w'.", path, err);
    }

    response.Assignment = &result;

    return 0, response, nil;
}
