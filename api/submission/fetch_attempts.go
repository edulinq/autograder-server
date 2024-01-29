package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
)

type FetchAttemptsRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleGrader

    TargetUser core.TargetUserSelfOrGrader `json:"target-email"`
}

type FetchAttemptsResponse struct {
    FoundUser bool `json:"found-user"`
    GradingResults []*model.GradingResult `json:"grading-results"`
}

func HandleFetchAttempts(request *FetchAttemptsRequest) (*FetchAttemptsResponse, *core.APIError) {
    response := FetchAttemptsResponse{
        FoundUser: false,
        GradingResults: make([]*model.GradingResult, 0),
    }

    if !request.TargetUser.Found {
        return &response, nil;
    }

    response.FoundUser = true;

    gradingResults, err := db.GetSubmissionAttempts(request.Assignment, request.TargetUser.Email);
    if err != nil {
        return nil, core.NewInternalError("-607", &request.APIRequestCourseUserContext, "Failed to get submission attempts.").
            Err(err).Add("email", request.TargetUser.Email);
    }

    response.GradingResults = gradingResults;

    return &response, nil;
}
