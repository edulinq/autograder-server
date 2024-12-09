package user

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type FetchUserPeekRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleStudent

	TargetUser       core.TargetCourseUserSelfOrGrader `json:"target-email"`
	TargetSubmission string                            `json:"target-submission"`
}

type FetchUserPeekResponse struct {
	FoundUser       bool               `json:"found-user"`
	FoundSubmission bool               `json:"found-submission"`
	GradingInfo     *model.GradingInfo `json:"submission-result"`
}

// Get a copy of the grading report for the specified submission. Does not submit a new submission.
func HandleFetchUserPeek(request *FetchUserPeekRequest) (*FetchUserPeekResponse, *core.APIError) {
	response := FetchUserPeekResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	submissionResult, err := db.GetSubmissionResult(request.Assignment, request.TargetUser.Email, request.TargetSubmission)
	if err != nil {
		return nil, core.NewInternalError("-601", &request.APIRequestCourseUserContext, "Failed to get submission result.").
			Err(err).Assignment(request.Assignment.GetID()).
			Add("target-user", request.TargetUser.Email).Add("submission", request.TargetSubmission)
	}

	if submissionResult == nil {
		return &response, nil
	}

	response.FoundSubmission = true
	response.GradingInfo = submissionResult

	return &response, nil
}
