package submission

import (
    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/model"
)

type PeekRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleStudent

    TargetUser core.TargetUserSelfOrGrader `json:"target-email"`
    TargetSubmission string `json:"target-submission"`
}

type PeekResponse struct {
    FoundUser bool `json:"found-user"`
    FoundSubmission bool `json:"found-submission"`
    GradingInfo *model.GradingInfo `json:"submission-result"`
}

func HandlePeek(request *PeekRequest) (*PeekResponse, *core.APIError) {
    response := PeekResponse{};

    if (!request.TargetUser.Found) {
        return &response, nil;
    }

    response.FoundUser = true;

    submissionResult, err := db.GetSubmissionResult(request.Assignment, request.TargetUser.Email, request.TargetSubmission);
    if (err != nil) {
        return nil, core.NewInternalError("-601", &request.APIRequestCourseUserContext, "Failed to get submission result.").
                Err(err).Assignment(request.Assignment.GetID()).
                Add("target-user", request.TargetUser.Email).Add("submission", request.TargetSubmission);
    }

    if (submissionResult == nil) {
        return &response, nil;
    }

    response.FoundSubmission = true;
    response.GradingInfo = submissionResult;

    return &response, nil;
}
