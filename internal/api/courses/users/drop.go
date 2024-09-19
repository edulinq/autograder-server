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

	if !checkDropPermissions(request) {
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

func checkDropPermissions(request *DropRequest) bool {
	// If the request is from a server admin or above, they can drop anyone from the course.
	if request.ServerUser.Role >= model.ServerRoleAdmin {
		return true
	}

	// Non-server admin can only drop users with lower course roles than themselves.
	if request.User.Role > request.TargetCourseUser.User.Role {
		return true
	}

	return false
}
