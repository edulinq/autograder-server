package assignments

import (
	"github.com/edulinq/autograder/internal/api/core"
)

type ListRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleOther
}

type ListResponse struct {
	Assignments []*core.AssignmentInfo `json:"assignments"`
}

func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
	response := ListResponse{
		Assignments: core.NewAssignmentInfos(request.Course.GetSortedAssignments()),
	}

	return &response, nil
}
