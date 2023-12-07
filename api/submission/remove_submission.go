package submission

import (
	"github.com/eriq-augustine/autograder/api/core"
	"github.com/eriq-augustine/autograder/db"
)

type RemoveSubmissionRequest struct {
	core.APIRequestAssignmentContext
	core.MinRoleGrader

	TargetUser core.TargetUserSelfOrGrader `json:"target-email"`
	TargetSubmission string `json:"target-submission"`
}

type RemoveSubmissionResponse struct {
	FoundUser bool `json:"found-user"`
	FoundSubmission bool `json:"found-submission"`
}

func HandleRemoveSubmission(request *RemoveSubmissionRequest) (*RemoveSubmissionResponse, *core.APIError){
	response := RemoveSubmissionResponse{};
	if (!request.TargetUser.Found) {
		return &response, nil;
	}

	response.FoundUser = true;
	err := db.RemoveSubmission(request.Assignment, request.TargetUser.Email, request.TargetSubmission)
	if (err != nil){
		response.FoundSubmission = false;
	} else {
		response.FoundSubmission = true;
	}
	
	return &response, nil;
}