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

// Authenticate as a user.
func HandleAuth(request *AuthRequest) (*AuthResponse, *core.APIError) {
	response := AuthResponse{true}

	return &response, nil
}
