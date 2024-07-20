package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type ListRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin
}

type ListResponse struct {
	Users []*core.ServerUserInfo `json:"users"`
}

func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	response := ListResponse{}

	users_map, err := db.GetServerUsers()
	if err != nil {
		return nil, core.NewBaseInternalError("-805", &request.APIRequest, "Failed to get server users from database.").Err(err)
	}

	users := make([]*model.ServerUser, 0, len(users_map))
	for _, user := range users_map {
		users = append(users, user)
	}

	info, err := core.NewServerUserInfos(users)
	if err != nil {
		return nil, core.NewBaseInternalError("-806", &request.APIRequest, "Failed to get server user infos.").Err(err)
	}

	response.Users = info

	return &response, nil
}
