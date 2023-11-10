package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/db"
)

type HistoryRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleStudent
    Users core.CourseUsers

    TargetUser core.TargetUserSelfOrGrader `json:"target-email"`
}

type HistoryResponse struct {
    FoundUser bool `json:"found-user"`
    History []*artifact.SubmissionHistoryItem `json:"history"`
}

func HandleHistory(request *HistoryRequest) (*HistoryResponse, *core.APIError) {
    response := HistoryResponse{
        FoundUser: false,
        History: make([]*artifact.SubmissionHistoryItem, 0),
    };

    if (!request.TargetUser.Found) {
        return &response, nil;
    }

    response.FoundUser = true;

    history, err := db.GetSubmissionHistory(request.Assignment, request.TargetUser.Email);
    if (err != nil) {
        return nil, core.NewInternalError("-407", &request.APIRequestCourseUserContext, "Failed to get submission history.").
                Err(err).Add("user", request.TargetUser.Email);
    }

    response.History = history;

    return &response, nil;
}
