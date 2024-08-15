package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type PassChangeRequest struct {
	core.APIRequestUserContext

	NewPass string `json:"new-pass"`
}

type PassChangeResponse struct {
	Success   bool `json:"success"`
	Duplicate bool `json:"duplicate"`
}

func HandlePassChange(request *PassChangeRequest) (*PassChangeResponse, *core.APIError) {
	success, err := request.ServerUser.SetPassword(request.NewPass)
	if err != nil {
		return nil, core.NewUsertContextInternalError("-805", &request.APIRequestUserContext,
			"Failed to set new password.").Err(err)
	}

	err = db.UpsertUser(request.ServerUser)
	if err != nil {
		return nil, core.NewUsertContextInternalError("-806", &request.APIRequestUserContext,
			"Failed to save user.").Err(err)
	}

	response := &PassChangeResponse{
		Success:   true,
		Duplicate: !success,
	}

	return response, nil
}
