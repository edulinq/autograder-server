package tokens

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type TokensCreateRequest struct {
	core.APIRequestUserContext

	Name string `json:"name"`
}

type TokensCreateResponse struct {
	TokenID        string `json:"token-id"`
	TokenCleartext string `json:"token-cleartext"`
}

// Create a new authentication token.
func HandleTokensCreate(request *TokensCreateRequest) (*TokensCreateResponse, *core.APIError) {
	token, cleartext, err := request.ServerUser.CreateRandomToken(request.Name, model.TokenSourceUser)
	if err != nil {
		return nil, core.NewUserContextInternalError("-801", &request.APIRequestUserContext,
			"Failed to create random user token.").Err(err)
	}

	err = db.UpsertUser(request.ServerUser)
	if err != nil {
		return nil, core.NewUserContextInternalError("-802", &request.APIRequestUserContext,
			"Failed to save user.").Err(err)
	}

	response := &TokensCreateResponse{
		TokenID:        token.ID,
		TokenCleartext: cleartext,
	}

	return response, nil
}
