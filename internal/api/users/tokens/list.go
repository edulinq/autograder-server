package tokens

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
)

type ListRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleUser

	TargetUser core.TargetServerUserSelfOrAdmin `json:"target-user"`
}

type ListResponse struct {
	FoundUser bool              `json:"found-user"`
	Tokens    []model.TokenInfo `json:"tokens"`
}

// List all the token information for a server user.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	response := ListResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundUser = true
	response.Tokens = make([]model.TokenInfo, 0, len(request.TargetUser.User.Tokens))

	for _, token := range request.TargetUser.User.Tokens {
		response.Tokens = append(response.Tokens, token.TokenInfo)
	}

	return &response, nil
}
