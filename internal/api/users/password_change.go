package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type PasswordChangeRequest struct {
	core.APIRequestUserContext

	NewPass string `json:"new-pass"`
}

type PasswordChangeResponse struct {
	Success   bool `json:"success"`
	Duplicate bool `json:"duplicate"`
}

// Change your password to the one provided.
func HandlePasswordChange(request *PasswordChangeRequest) (*PasswordChangeResponse, *core.APIError) {
	success, err := request.ServerUser.SetPassword(request.NewPass)
	if err != nil {
		return nil, core.NewUserContextInternalError("-805", &request.APIRequestUserContext,
			"Failed to set new password.").Err(err)
	}

	err = db.UpsertUser(request.ServerUser)
	if err != nil {
		return nil, core.NewUserContextInternalError("-806", &request.APIRequestUserContext,
			"Failed to save user.").Err(err)
	}

	response := &PasswordChangeResponse{
		Success:   true,
		Duplicate: !success,
	}

	return response, nil
}
