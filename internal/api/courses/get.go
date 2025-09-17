package courses

import (
	"github.com/edulinq/autograder/internal/api/core"
)

type GetRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleOther
}

type GetResponse struct {
	Found  bool             `json:"found"`
	Course *core.CourseInfo `json:"course"`
}

// Get information about a course.
func HandleGet(request *GetRequest) (*GetResponse, *core.APIError) {
	response := GetResponse{
		Found:  true,
		Course: core.NewCourseInfo(request.Course),
	}

	return &response, nil
}
