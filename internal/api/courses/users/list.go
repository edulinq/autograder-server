package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type ListRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	Users []model.CourseUserReference `json:"users"`
}

type ListResponse struct {
	Users  []*core.CourseUserInfo `json:"users"`
	Errors map[string]string      `json:"errors,omitempty"`
}

// List the users in the course.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	// Default to listing all users in the course.
	if len(request.Users) == 0 {
		request.Users = []model.CourseUserReference{"*"}
	}

	reference, err := model.ResolveCourseUserReferences(request.Course, request.Users)
	if err != nil {
		return nil, core.NewInternalError("-635", request, "Failed to parse course user references.").Err(err)
	}

	errors := make(map[string]string, len(userErrors))

	for user, err := range userErrors {
		errors[user] = err.Error()

		log.Warn("Failed to parse user reference.", err, log.NewAttr("reference", user))
	}

	users, err := db.ResolveCourseUsers(request.Course, reference)
	if err != nil {
		return nil, core.NewInternalError("-636", request, "Failed to resolve course users.").Err(err)
	}

	response := ListResponse{
		Users:  core.NewCourseUserInfos(users),
		Errors: errors,
	}

	return &response, nil
}
