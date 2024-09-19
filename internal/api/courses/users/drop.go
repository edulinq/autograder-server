package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type DropRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	TargetCourseUser core.TargetCourseUser `json:"target-email"`
}

type DropResponse struct {
	FoundUser bool `json:"found-user"`
}

func HandleDrop(request *DropRequest) (*DropResponse, *core.APIError) {
	response := DropResponse{}

	if !request.TargetCourseUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	if (request.ServerUser.Role < model.ServerRoleAdmin) && (request.TargetCourseUser.User.Role >= request.User.Role) {
		return nil, core.NewBadCoursePermissionsError("-612", &request.APIRequestCourseUserContext, request.TargetCourseUser.User.Role,
			"Cannot drop a user with an equal or higher role.").Add("target-course-user", request.TargetCourseUser.User.Email)
	}

	_, _, err := db.RemoveUserFromCourse(request.Course, request.TargetCourseUser.User.Email)
	if err != nil {
		return nil, core.NewInternalError("-613", &request.APIRequestCourseUserContext,
			"Failed to drop user.").Err(err).Add("target-course-user", request.TargetCourseUser.User.Email)
	}

	return &response, nil
}
