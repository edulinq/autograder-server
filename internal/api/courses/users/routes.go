package users

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`courses/users/drop`, HandleDrop, "Drop a user from the course."),
	core.NewAPIRoute(`courses/users/enroll`, HandleEnroll, "Enroll one or more users to the course."),
	core.NewAPIRoute(`courses/users/get`, HandleGet, "Get the information for a course user."),
	core.NewAPIRoute(`courses/users/list`, HandleList, "List the users in the course."),
}

func GetRoutes() *[]core.Route {
	return &routes
}
