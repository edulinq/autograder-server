package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type ListRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	Users []model.CourseUserReferenceInput `json:"users"`
}

type ListResponse struct {
	Users []*core.CourseUserInfo `json:"users"`
}

// List the users in the course.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	// Default to listing all users in the course.
	if len(request.Users) == 0 {
		request.Users = []model.CourseUserReferenceInput{"*"}
	}

	reference, userErr, err := db.ParseCourseUserReference(request.Course, request.Users)
	if err != nil {
		return nil, core.NewInternalError("-635", request, "Failed to parse course user references.").Err(err)
	}

	if userErr != nil {
		return nil, core.NewBadRequestError("-636", request, "Invalid course user references.").Err(userErr)
	}

	users, err := db.ResolveCourseUsers(request.Course, reference)
	if err != nil {
		return nil, core.NewInternalError("-637", request, "Failed to resolve course users.").Err(err)
	}

	response := ListResponse{core.NewCourseUserInfos(users)}

	return &response, nil
}
