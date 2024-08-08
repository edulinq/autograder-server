package courses

// All routes handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments"
	"github.com/edulinq/autograder/internal/api/courses/users"
)

func GetRoutes() *[]*core.Route {
	routes := make([]*core.Route, 0)

	routes = append(routes, *(assignments.GetRoutes())...)
	routes = append(routes, *(users.GetRoutes())...)

	return &routes
}
