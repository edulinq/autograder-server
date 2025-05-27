package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type RemoveRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	TargetUser core.TargetServerUser `json:"target-email" required:""`
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
		return nil, core.NewPermissionsError("-811", request, request.TargetUser.User.Role, request.ServerUser.Role,
			"Cannot remove a user with an equal or higher role.").Add("target-user", request.TargetUser.User.Email)
	}

	_, err := db.DeleteUser(request.TargetUser.Email)
	if err != nil {
		return nil, core.NewInternalError("-812", request,
			"Failed to remove user.").Err(err).Add("target-user", request.TargetUser.User.Email)
	}

	return &response, nil
}
