package web

import (
    "fmt"
    "net/http"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

type HistoryRequest struct {
    model.BaseAPIRequest
    Assignment string `json:"assignment"`
}

type HistoryResponse struct {
    History []*model.SubmissionSummary `json:"history"`
}

func (this *HistoryRequest) String() string {
    return util.BaseString(this);
}

func NewHistoryRequest(request *http.Request) (*HistoryRequest, *model.APIResponse, error) {
    var apiRequest HistoryRequest;
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

func (this *HistoryRequest) Close() error {
    return nil;
}

func (this *HistoryRequest) Clean() error {
    var err error;
    this.Assignment, err = model.ValidateID(this.Assignment);
    if (err != nil) {
        return fmt.Errorf("Could not clean HistoryRequest assignment ID ('%s'): '%w'.", this.Assignment, err);
    }

    return nil;
}

func handleHistory(request *HistoryRequest) (int, any, error) {
    assignment := grader.GetAssignment(request.Course, request.Assignment);
    if (assignment == nil) {
        return http.StatusBadRequest, fmt.Sprintf("Could not find assignment ('%s') for course ('%s').", request.Assignment, request.Course,), nil;
    }

    paths, err := assignment.GetSubmissionSummaries(request.User);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed tp get submission summaries: '%w'.", err);
    }

    response := HistoryResponse{
        History: make([]*model.SubmissionSummary, 0, len(paths)),
    };

    for _, path := range paths {
        summary := model.SubmissionSummary{};
        err = util.JSONFromFile(path, &summary);
        if (err != nil) {
            return 0, nil, fmt.Errorf("Failed to deserialize submission summary '%s': '%w'.", path, err);
        }

        response.History = append(response.History, &summary);
    }

    return 0, response, nil;
}
