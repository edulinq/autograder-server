package admin

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/procedures/courses"
)

type UpdateRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin
}

type UpdateResponse struct {
	Result *courses.CourseUpsertResult `json:"result"`
}

// Update an existing course.
func HandleUpdate(request *UpdateRequest) (*UpdateResponse, *core.APIError) {
	options := courses.CourseUpsertOptions{
		ContextUser: request.ServerUser,
	}

	result, err := courses.UpdateFromLocalSource(request.Course, options)
	if err != nil {
		return nil, core.NewInternalError("-611", request,
			"Failed to update course.").Err(err)
	}

	return &UpdateResponse{result}, nil
}
