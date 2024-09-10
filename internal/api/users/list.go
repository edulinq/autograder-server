package users

import (
	"slices"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
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
		return nil, core.NewUsertContextInternalError("-813", &request.APIRequestUserContext,
			"Failed to get server users from database.").Err(err)
	}

	infos := make([]*core.ServerUserInfo, 0, len(usersMap))
	for _, user := range usersMap {
		info, err := core.NewServerUserInfo(user)
		if err != nil {
			return nil, core.NewUsertContextInternalError("-814", &request.APIRequestUserContext,
				"Failed to get server user info.").Err(err)
		}

		infos = append(infos, info)
	}

	slices.SortFunc(infos, core.CompareServerUserInfoPointer)

	response := ListResponse{infos}

	return &response, nil
}
