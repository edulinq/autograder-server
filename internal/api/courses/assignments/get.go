package assignments

import (
	"github.com/edulinq/autograder/internal/api/core"
)

type GetRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleOther
}

type GetResponse struct {
	Assignment *core.AssignmentInfo `json:"assignment"`
}

// Get the information for a course assignment.
func HandleGet(request *GetRequest) (*GetResponse, *core.APIError) {
	response := GetResponse{
		Assignment: core.NewAssignmentInfo(request.Assignment),
	}

	return &response, nil
}
