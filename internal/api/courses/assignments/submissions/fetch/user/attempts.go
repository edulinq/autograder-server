package user

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type FetchUserAttemptsRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	TargetUser core.TargetCourseUserSelfOrGrader `json:"target-email"`
}

type FetchUserAttemptsResponse struct {
	FoundUser      bool                   `json:"found-user"`
	GradingResults []*model.GradingResult `json:"grading-results"`
}

// Get all submission attempts made by a user along with all grading information.
func HandleFetchUserAttempts(request *FetchUserAttemptsRequest) (*FetchUserAttemptsResponse, *core.APIError) {
	response := FetchUserAttemptsResponse{
		FoundUser:      false,
		GradingResults: make([]*model.GradingResult, 0),
	}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	gradingResults, err := db.GetSubmissionAttempts(request.Assignment, request.TargetUser.Email)
	if err != nil {
		return nil, core.NewInternalError("-607", request, "Failed to get submission attempts.").
			Err(err).Add("target-user", request.TargetUser.Email)
	}

	response.GradingResults = gradingResults

	return &response, nil
}
