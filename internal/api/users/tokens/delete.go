package tokens

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type DeleteRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleUser

	TargetUser core.TargetServerUserSelfOrAdmin `json:"target-user"`

	TokenID core.NonEmptyString `json:"token-id" required:""`
}

type DeleteResponse struct {
	FoundUser  bool `json:"found-user"`
	FoundToken bool `json:"found-token"`
}

// Delete an authentication token.
func HandleDelete(request *DeleteRequest) (*DeleteResponse, *core.APIError) {
	response := DeleteResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	found, err := db.DeleteUserToken(request.TargetUser.User.Email, string(request.TokenID))
	if err != nil {
		return nil, core.NewInternalError("-803", request,
			"Failed to delete user token.").Err(err)
	}

	response.FoundToken = found

	return &response, nil
}
