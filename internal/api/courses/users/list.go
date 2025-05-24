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

	References []model.CourseUserReference `json:"references"`
}

type ListResponse struct {
	Users  []*core.CourseUserInfo `json:"users"`
	Errors map[string]string      `json:"errors,omitempty"`
}

// List the users in the course.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	// Default to listing all users in the course.
	if len(request.References) == 0 {
		request.References = []model.CourseUserReference{"*"}
	}

	reference, userErrors := model.ResolveCourseUserReferences(request.References)

	errors := make(map[string]string, len(userErrors))

	for reference, err := range userErrors {
		errors[reference] = err.Error()

		log.Warn("Failed to parse user reference.", err, log.NewAttr("reference", reference))
	}

	users := model.ResolveCourseUsers(request.Users, reference)

	response := ListResponse{
		Users:  core.NewCourseUserInfos(users),
		Errors: errors,
	}

	return &response, nil
}
