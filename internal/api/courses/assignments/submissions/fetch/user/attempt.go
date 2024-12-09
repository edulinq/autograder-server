package user

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type FetchUserAttemptRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleStudent

	TargetUser       core.TargetCourseUserSelfOrGrader `json:"target-email"`
	TargetSubmission string                            `json:"target-submission"`
}

type FetchUserAttemptResponse struct {
	FoundUser       bool                 `json:"found-user"`
	FoundSubmission bool                 `json:"found-submission"`
	GradingResult   *model.GradingResult `json:"grading-result"`
}

// Get a submission along with all grading information.
func HandleFetchUserAttempt(request *FetchUserAttemptRequest) (*FetchUserAttemptResponse, *core.APIError) {
	response := FetchUserAttemptResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	gradingResult, err := db.GetSubmissionContents(request.Assignment, request.TargetUser.Email, request.TargetSubmission)
	if err != nil {
		return nil, core.NewInternalError("-604", &request.APIRequestCourseUserContext, "Failed to get submission contents.").
			Err(err).Assignment(request.Assignment.GetID()).
			Add("target-user", request.TargetUser.Email).Add("submission", request.TargetSubmission)
	}

	if gradingResult == nil {
		return &response, nil
	}

	response.FoundSubmission = true
	response.GradingResult = gradingResult

	return &response, nil
}
