package users

import (
	"maps"
	"slices"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type ListRequest struct {
	core.APIRequest
	core.MinServerRoleAdmin
}

type ListResponse struct {
	Users *core.ServerUserInfos `json:"users"`
}

func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	response := ListResponse{}

	users_map, err := db.GerServerUsers()
	if err != nil {
		return nil, core.NewBaseInternalError("-805", &request.APIRequest, "Failed to get server users from database.").Err(err)
	}

	users := slices.Collect(maps.Values(users_map))

	info, err := core.NewServerUserInfos(users)
	if err != nil {
		return nil, core.NewBaseInternalError("-806", &request.APIRequest, "Failed to get server user infos.").Err(err)
	}

	response.Users = info

	return &response, nil
}
