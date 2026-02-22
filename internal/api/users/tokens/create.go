package tokens

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type CreateRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleUser

	TargetUser core.TargetServerUserSelfOrAdmin `json:"target-user"`

	Name string `json:"name"`
}

type CreateResponse struct {
	FoundUser      bool             `json:"found-user"`
	TokenInfo      *model.TokenInfo `json:"token-info"`
	TokenCleartext string           `json:"token-cleartext"`
}

// Create a new authentication token.
func HandleCreate(request *CreateRequest) (*CreateResponse, *core.APIError) {
	response := CreateResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	token, cleartext, err := request.TargetUser.User.CreateRandomToken(request.Name, model.TokenSourceUser)
	if err != nil {
		return nil, core.NewInternalError("-801", request,
			"Failed to create random user token.").Err(err)
	}

	err = db.UpsertUser(request.TargetUser.User)
	if err != nil {
		return nil, core.NewInternalError("-802", request,
			"Failed to save user.").Err(err)
	}

	response.TokenInfo = &token.TokenInfo
	response.TokenCleartext = cleartext

	return &response, nil
}
