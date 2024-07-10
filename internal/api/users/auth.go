package users

import (
	"github.com/edulinq/autograder/internal/api/core"
)

type AuthRequest struct {
	core.APIRequestUserContext
}

type AuthResponse struct {
	Success bool `json:"success"`
}

func HandleAuth(request *AuthRequest) (*AuthResponse, *core.APIError) {
	response := AuthResponse{}
	response.Success = true

	return &response, nil
}
