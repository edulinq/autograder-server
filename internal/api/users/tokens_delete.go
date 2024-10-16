package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type TokensDeleteRequest struct {
	core.APIRequestUserContext

	TokenID core.NonEmptyString `json:"token-id"`
}

type TokensDeleteResponse struct {
	Found bool `json:"found"`
}

// Delete an authentication token.
func HandleTokensDelete(request *TokensDeleteRequest) (*TokensDeleteResponse, *core.APIError) {
	found, err := db.DeleteUserToken(request.ServerUser.Email, string(request.TokenID))
	if err != nil {
		return nil, core.NewUserContextInternalError("-803", &request.APIRequestUserContext,
			"Failed to delete user token.").Err(err)
	}

	response := &TokensDeleteResponse{
		Found: found,
	}

	return response, nil
}
