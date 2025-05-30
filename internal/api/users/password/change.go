package password

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type PasswordChangeRequest struct {
	core.APIRequestUserContext

	NewPass string `json:"new-pass" required:""`
}

type PasswordChangeResponse struct {
	Success   bool `json:"success"`
	Duplicate bool `json:"duplicate"`
}

// Change your password to the one provided.
func HandlePasswordChange(request *PasswordChangeRequest) (*PasswordChangeResponse, *core.APIError) {
	success, err := request.ServerUser.SetPassword(request.NewPass)
	if err != nil {
		return nil, core.NewInternalError("-805", request,
			"Failed to set new password.").Err(err)
	}

	err = db.UpsertUser(request.ServerUser)
	if err != nil {
		return nil, core.NewInternalError("-806", request,
			"Failed to save user.").Err(err)
	}

	response := &PasswordChangeResponse{
		Success:   true,
		Duplicate: !success,
	}

	return response, nil
}
