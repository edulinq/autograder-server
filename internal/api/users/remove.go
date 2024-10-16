package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type RemoveRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	TargetUser core.TargetServerUser `json:"target-email"`
}

type RemoveResponse struct {
	FoundUser bool `json:"found-user"`
}

// Remove a user from the server.
func HandleRemove(request *RemoveRequest) (*RemoveResponse, *core.APIError) {
	response := RemoveResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	if request.TargetUser.User.Role >= request.ServerUser.Role {
		return nil, core.NewBadServerPermissionsError("-811", &request.APIRequestUserContext, request.TargetUser.User.Role,
			"Cannot remove a user with an equal or higher role.").Add("target-user", request.TargetUser.User.Email)
	}

	_, err := db.DeleteUser(request.TargetUser.Email)
	if err != nil {
		return nil, core.NewUserContextInternalError("-812", &request.APIRequestUserContext,
			"Failed to remove user.").Err(err).Add("target-user", request.TargetUser.User.Email)
	}

	return &response, nil
}
