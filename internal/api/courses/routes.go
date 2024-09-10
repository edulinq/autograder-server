package courses

// All routes handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/admin"
	"github.com/edulinq/autograder/internal/api/courses/assignments"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
)

func GetRoutes() *[]*core.Route {
	routes := make([]*core.Route, 0)

	routes = append(routes, *(admin.GetRoutes())...)
	routes = append(routes, *(assignments.GetRoutes())...)
	routes = append(routes, *(submissions.GetRoutes())...)

	return &routes
}
