package courses

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type ListRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin
}

type ListResponse struct {
	Courses []*core.CourseInfo `json:"courses"`
}

// List the courses on the server.
func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	courses, err := db.GetCourses()
	if err != nil {
		return nil, core.NewInternalError("-641", request, "Failed to get courses from database.").Err(err)
	}

	return &ListResponse{core.NewCourseInfosFromMap(courses)}, nil
}
