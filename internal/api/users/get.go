package users

import (
	"github.com/edulinq/autograder/internal/api/core"
)

type GetRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleUser

	TargetUser core.TargetServerUserSelfOrAdmin `json:"target-email"`
}

type GetResponse struct {
	Found bool                 `json:"found"`
	User  *core.ServerUserInfo `json:"user"`
}

func HandleGet(request *GetRequest) (*GetResponse, *core.APIError) {
	response := GetResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.Found = true

	info, err := core.NewServerUserInfo(request.TargetUser.User)
	if err != nil {
		return nil, core.NewUserContextInternalError("-804", &request.APIRequestUserContext,
			"Failed to get server user info.").Err(err)
	}

	response.User = info

	return &response, nil
}
