package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/util"
)

type HistoryRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleStudent
    Users core.CourseUsers

    TargetUser core.TargetUserSelfOrGrader `json:"target-email"`
}

type HistoryResponse struct {
    FoundUser bool `json:"found-user"`
    History []*artifact.SubmissionSummary `json:"history"`
}

func HandleHistory(request *HistoryRequest) (*HistoryResponse, *core.APIError) {
    response := HistoryResponse{
        FoundUser: false,
        History: make([]*artifact.SubmissionSummary, 0),
    };

    if (!request.TargetUser.Found) {
        return &response, nil;
    }

    response.FoundUser = true;

    paths, err := request.Assignment.GetSubmissionSummaries(request.TargetUser.Email);
    if (err != nil) {
        return nil, core.NewInternalError("-407", &request.APIRequestCourseUserContext, "Failed to get submission summaries.").Err(err);
    }

    for _, path := range paths {
        summary := artifact.SubmissionSummary{};
        err = util.JSONFromFile(path, &summary);
        if (err != nil) {
            return nil, core.NewInternalError("-408", &request.APIRequestCourseUserContext,
                    "Failed to deserialize submission summary.").Err(err).Add("path", path);
        }

        response.History = append(response.History, &summary);
    }

    return &response, nil;
}
