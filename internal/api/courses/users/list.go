package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type ListRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader
	Users core.CourseUsers `json:"-"`

	TargetUsers []model.CourseUserReference `json:"target-users"`
}

type ListResponse struct {
	Users  []*core.CourseUserInfo `json:"users"`
	Errors map[string]string      `json:"errors,omitempty,omitzero"`
}

// List the users in the course.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	// Default to listing all users in the course.
	if len(request.TargetUsers) == 0 {
		request.TargetUsers = []model.CourseUserReference{"*"}
	}

	reference, userErrors := model.ParseCourseUserReferences(request.TargetUsers)

	errors := make(map[string]string, len(userErrors))

	for reference, err := range userErrors {
		errors[reference] = err.Error()

		log.Warn("Failed to parse user reference.", err, log.NewAttr("reference", reference))
	}

	if len(userErrors) != 0 {
		return &ListResponse{
			Users:  nil,
			Errors: errors,
		}, nil
	}

	users := model.ResolveCourseUsers(request.Users, reference)

	response := ListResponse{
		Users:  core.NewCourseUserInfos(users),
		Errors: nil,
	}

	return &response, nil
}
