package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/db"
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
    SubmissionResult *artifact.GradedAssignment `json:"submission-result"`
}

func HandlePeek(request *PeekRequest) (*PeekResponse, *core.APIError) {
    response := PeekResponse{};

    if (!request.TargetUser.Found) {
        return &response, nil;
    }

    response.FoundUser = true;

    submissionResult, err := db.GetSubmissionResult(request.Assignment, request.TargetUser.Email, request.TargetSubmission);
    if (err != nil) {
        return nil, core.NewInternalError("-402", &request.APIRequestCourseUserContext, "Failed to get submission result.").
                Err(err).Add("user", request.TargetUser.Email).Add("submission", request.TargetSubmission);
    }

    if (submissionResult == nil) {
        return &response, nil;
    }

    response.FoundSubmission = true;
    response.SubmissionResult = submissionResult;

    return &response, nil;
}
