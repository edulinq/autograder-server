package courses

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments"
)

func GetRoutes() *[]*core.Route {
	routes := make([]*core.Route, 0)

	routes = append(routes, *(assignments.GetRoutes())...)

	return &routes
}
