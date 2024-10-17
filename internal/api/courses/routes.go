package courses

// All routes handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/admin"
	"github.com/edulinq/autograder/internal/api/courses/assignments"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
	"github.com/edulinq/autograder/internal/api/courses/upsert"
	"github.com/edulinq/autograder/internal/api/courses/users"
)

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, *(admin.GetRoutes())...)
	routes = append(routes, *(assignments.GetRoutes())...)
	routes = append(routes, *(submissions.GetRoutes())...)
	routes = append(routes, *(upsert.GetRoutes())...)
	routes = append(routes, *(users.GetRoutes())...)

	return &routes
}
