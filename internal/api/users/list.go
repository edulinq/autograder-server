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

// List the users on the server.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	usersMap, err := db.GetServerUsers()
	if err != nil {
		return nil, core.NewInternalError("-813", request,
			"Failed to get server users from database.").Err(err)
	}

	infos := make([]*core.ServerUserInfo, 0, len(usersMap))
	for _, user := range usersMap {
		infos = append(infos, core.NewServerUserInfo(user))
	}

	slices.SortFunc(infos, core.CompareServerUserInfoPointer)

	response := ListResponse{infos}

	return &response, nil
}
