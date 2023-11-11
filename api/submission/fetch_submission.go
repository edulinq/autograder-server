package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/db"
)

type FetchSubmissionRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleStudent

    TargetUser core.TargetUserSelfOrGrader `json:"target-email"`
    TargetSubmission string `json:"target-submission"`
}

type FetchSubmissionResponse struct {
    FoundUser bool `json:"found-user"`
    FoundSubmission bool `json:"found-submission"`
    GradingResult *artifact.GradingResult `json:"grading-result"`
}

func HandleFetchSubmission(request *FetchSubmissionRequest) (*FetchSubmissionResponse, *core.APIError) {
    response := FetchSubmissionResponse{};

    if (!request.TargetUser.Found) {
        return &response, nil;
    }

    response.FoundUser = true;

    gradingResult, err := db.GetSubmissionContents(request.Assignment, request.TargetUser.Email, request.TargetSubmission);
    if (err != nil) {
        return nil, core.NewInternalError("-409", &request.APIRequestCourseUserContext, "Failed to get submission contents.").
                Err(err).Add("email", request.TargetUser.Email).Add("submission", request.TargetSubmission);
    }

    if (gradingResult == nil) {
        return &response, nil;
    }

    response.FoundSubmission = true;
    response.GradingResult = gradingResult;

    return &response, nil;
}
