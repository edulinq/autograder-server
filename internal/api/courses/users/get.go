package users

import (
	"github.com/edulinq/autograder/internal/api/core"
)

type GetRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleOther

	TargetCourseUser core.TargetCourseUserSelfOrGrader `json:"target-email"`
}

type GetResponse struct {
	Found bool                 `json:"found"`
	User  *core.CourseUserInfo `json:"user"`
}

// Get the information for a course user.
func HandleGet(request *GetRequest) (*GetResponse, *core.APIError) {
	response := GetResponse{}

	if !request.TargetCourseUser.Found {
		return &response, nil
	}

	response.Found = true
	response.User = core.NewCourseUserInfo(request.TargetCourseUser.User)

	return &response, nil
}
