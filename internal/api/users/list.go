package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type ListRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	TargetUsers []model.ServerUserReference `json:"target-users"`
}

type ListResponse struct {
	Users []*core.ServerUserInfo `json:"users"`
}

// List the users on the server.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	// Default to listing all users in the server.
	if len(request.TargetUsers) == 0 {
		request.TargetUsers = model.NewAllServerUserReference()
	}

	courses, err := db.GetCourses()
	if err != nil {
		return nil, core.NewInternalError("-814", request, "Failed to get courses from database.").Err(err)
	}

	reference, err := model.ParseServerUserReferences(request.TargetUsers, courses)
	if err != nil {
		return nil, core.NewBadRequestError("-815", request, "Failed to parse target users.").Err(err)
	}

	usersMap, err := db.GetServerUsers()
	if err != nil {
		return nil, core.NewInternalError("-813", request,
			"Failed to get server users from database.").Err(err)
	}

	users := model.ResolveServerUsers(usersMap, reference)

	response := ListResponse{core.NewServerUserInfos(users)}

	return &response, nil
}
