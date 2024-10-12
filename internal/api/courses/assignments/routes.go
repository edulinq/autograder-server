package assignments

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`courses/assignments/get`, HandleGet, "Get the information for a course assignment."),
	core.NewAPIRoute(`courses/assignments/list`, HandleList, "List the assignments in the course."),
}

func GetRoutes() *[]core.Route {
	return &routes
}
