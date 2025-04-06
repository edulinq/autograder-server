package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

type GetRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleUser

	TargetUser core.TargetServerUserSelfOrAdmin `json:"target-email"`
}

type GetResponse struct {
	Found   bool                        `json:"found"`
	User    *core.ServerUserInfo        `json:"user"`
	Courses map[string]*core.CourseInfo `json:"courses"`
}

// Get the information for a server user.
func HandleGet(request *GetRequest) (*GetResponse, *core.APIError) {
	response := GetResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.Found = true
	response.User = core.NewServerUserInfo(request.TargetUser.User)

	response.Courses = make(map[string]*core.CourseInfo, len(request.TargetUser.User.CourseInfo))
	for courseID, _ := range request.TargetUser.User.CourseInfo {
		course, err := db.GetCourse(courseID)
		if err != nil {
			return nil, core.NewInternalError("-804", request,
				"Failed to get user's course.").Err(err).Course(courseID)
		}

		if course != nil {
			response.Courses[courseID] = core.NewCourseInfo(course)
		}
	}

	return &response, nil
}
