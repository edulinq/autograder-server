package status

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
)

type ListRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleOther
}

type ListResponse struct {
	Statuses map[string]*model.CourseStatus `json:"statuses"`
}

// List all the statuses of a course.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	return &ListResponse{Statuses: request.Course.Statuses}, nil
}
