package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
)

type ListRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader
	Users core.CourseUsers `json:"-"`

	TargetUsers []model.CourseUserReference `json:"target-users"`
}

type ListResponse struct {
	Users []*core.CourseUserInfo `json:"users"`
}

// List the users in the course.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	// Default to listing all users in the course.
	if len(request.TargetUsers) == 0 {
		request.TargetUsers = []model.CourseUserReference{"*"}
	}

	reference, err := model.ParseCourseUserReferences(request.TargetUsers)
	if err != nil {
		return nil, core.NewBadRequestError("-635", request, "Failed to parse target users.").Err(err)
	}

	users := model.ResolveCourseUsers(request.Users, reference)

	response := ListResponse{core.NewCourseUserInfos(users)}

	return &response, nil
}
