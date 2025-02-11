package api

// All routes handled by the server.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses"
	"github.com/edulinq/autograder/internal/api/lms"
	"github.com/edulinq/autograder/internal/api/logs"
	"github.com/edulinq/autograder/internal/api/metadata"
	"github.com/edulinq/autograder/internal/api/static"
	"github.com/edulinq/autograder/internal/api/stats"
	"github.com/edulinq/autograder/internal/api/system"
	"github.com/edulinq/autograder/internal/api/users"
)

var baseRoutes = []core.Route{
	core.NewRedirect("GET", ``, `/static/index.html`),
	core.NewRedirect("GET", `/`, `/static/index.html`),
	core.NewRedirect("GET", `/index.html`, `/static/index.html`),

	core.NewBaseRoute("GET", `/static`, static.Handle),
	core.NewBaseRoute("GET", `/static/.*`, static.Handle),
}

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, baseRoutes...)
	routes = append(routes, *(courses.GetRoutes())...)
	routes = append(routes, *(lms.GetRoutes())...)
	routes = append(routes, *(logs.GetRoutes())...)
	routes = append(routes, *(metadata.GetRoutes())...)
	routes = append(routes, *(stats.GetRoutes())...)
	routes = append(routes, *(system.GetRoutes())...)
	routes = append(routes, *(users.GetRoutes())...)

	return &routes
}
