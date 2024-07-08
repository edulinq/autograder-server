package users

import (
	"fmt"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	// "github.com/edulinq/autograder/internal/util"
)

type AuthRequest struct {
	core.APIRequest
	core.MinServerRoleUser

	TargetUser core.NonEmptyString `json:"target-email"`
	TargetPass core.NonEmptyString `json:"target-pass"`
}

type AuthResponse struct {
	FoundUser   bool `json:"found-user"`
	AuthSuccess bool `json:"auth-success"`
}

func HandleAuth(request *AuthRequest) (*AuthResponse, *core.APIError) {
	response := AuthResponse{}

	user, err := db.GetServerUser(string(request.TargetUser), false)
	if err != nil {
		return nil, core.NewBadRequestError("-830", &request.APIRequest, "Error getting user").Err(err)
	}

	if user == nil {
		return &response, nil
	}

	fmt.Printf("User email: '%s', User pass: '%s'.\n", string(request.TargetUser), string(request.TargetPass))
	response.FoundUser = true
	response.AuthSuccess, err = user.Auth(string(request.TargetPass))
	if err != nil {
		return nil, core.NewBadRequestError("-831", &request.APIRequest, "Error during Auth").Err(err)
	}

	return &response, nil
}
