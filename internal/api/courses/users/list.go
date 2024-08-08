package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
)

type ListRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader
	Users core.CourseUsers `json:"-"`
}

type ListResponse struct {
	Users []*core.CourseUserInfo `json:"users"`
}

func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	users := make([]*model.CourseUser, 0, len(request.Users))
	for _, user := range request.Users {
		users = append(users, user)
	}

	infos := core.NewCourseUserInfos(users)

	response := ListResponse{infos}

	return &response, nil
}
