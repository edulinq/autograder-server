package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type AuthRequest struct {
	core.APIRequest

	TargetUser string `json:"target-email"`
	TargetPass string `json:"target-pass"`
}

type AuthResponse struct {
	FoundUser   bool `json:"found-user"`
	AuthSuccess bool `json:"auth-success"`
}

func HandleAuth(request *AuthRequest) (*AuthResponse, *core.APIError) {
	response := AuthResponse{}

	user, err := db.GetServerUser(request.TargetUser, true)
	if err != nil {
		return nil, core.NewBadRequestError("-012", &request.APIRequest, "Cannot Get User").Err(err)
	}

	if user == nil {
		return &response, nil
	}

	response.FoundUser = true
	response.AuthSuccess, err = user.Auth(request.TargetPass)
	if err != nil {
		return nil, core.NewBadRequestError("-037", &request.APIRequest, "User auth failed.").Err(err)
	}

	return &response, nil
}
