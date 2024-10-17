package submissions

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type RemoveRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	TargetUser       core.TargetCourseUserSelfOrGrader `json:"target-email"`
	TargetSubmission string                            `json:"target-submission"`
}

type RemoveResponse struct {
	FoundUser       bool `json:"found-user"`
	FoundSubmission bool `json:"found-submission"`
}

// Remove a specified submission. Defaults to the most recent submission.
func HandleRemove(request *RemoveRequest) (*RemoveResponse, *core.APIError) {
	response := RemoveResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	doesExist, err := db.RemoveSubmission(request.Assignment, request.TargetUser.Email, request.TargetSubmission)
	if err != nil {
		return nil, core.NewInternalError("-606", &request.APIRequestCourseUserContext, "Failed to remove the submission.").
			Err(err).Assignment(request.Assignment.GetID()).
			Add("target-user", request.TargetUser.Email).Add("submission", request.TargetSubmission)
	}

	response.FoundSubmission = doesExist

	return &response, nil
}
