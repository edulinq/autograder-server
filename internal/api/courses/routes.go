package courses

// All routes handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/admin"
	"github.com/edulinq/autograder/internal/api/courses/assignments"
	"github.com/edulinq/autograder/internal/api/courses/lms"
	"github.com/edulinq/autograder/internal/api/courses/stats"
	"github.com/edulinq/autograder/internal/api/courses/upsert"
	"github.com/edulinq/autograder/internal/api/courses/users"
)

var baseRoutes []core.Route = []core.Route{
	core.MustNewAPIRoute(`courses/get`, HandleGet),
}

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, baseRoutes...)
	routes = append(routes, *(admin.GetRoutes())...)
	routes = append(routes, *(assignments.GetRoutes())...)
	routes = append(routes, *(lms.GetRoutes())...)
	routes = append(routes, *(stats.GetRoutes())...)
	routes = append(routes, *(upsert.GetRoutes())...)
	routes = append(routes, *(users.GetRoutes())...)

	return &routes
}
