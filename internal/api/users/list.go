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
	usersMap, err := db.GetServerUsers()
	if err != nil {
		return nil, core.NewUsertContextInternalError("-811", &request.APIRequestUserContext,
			"Failed to get server users from database.").Err(err)
	}

	users := make([]*model.ServerUser, 0, len(usersMap))
	for _, user := range usersMap {
		users = append(users, user)
	}

	infos, err := core.NewServerUserInfos(users)
	if err != nil {
		return nil, core.NewUsertContextInternalError("-812", &request.APIRequestUserContext,
			"Failed to get server user infos.").Err(err)
	}

	response := ListResponse{infos}

	return &response, nil
}
