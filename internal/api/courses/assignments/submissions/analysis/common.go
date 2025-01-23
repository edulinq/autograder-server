package analysis

import (
	"github.com/edulinq/autograder/internal/model"
)

func checkPermissions(user *model.ServerUser, courses []string) bool {
	// Admins can do whatever they want.
	if user.Role >= model.ServerRoleAdmin {
		return true
	}

	// Regular server users need to be at least an admin in every course they are making requests for.
	for _, course := range courses {
		if user.GetCourseRole(course) < model.CourseRoleAdmin {
			return false
		}
	}

	return true
}
