package api

// All routes handled by the server.

import (
	"github.com/edulinq/autograder/internal/api/admin"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses"
	"github.com/edulinq/autograder/internal/api/lms"
	"github.com/edulinq/autograder/internal/api/users"
)

var baseRoutes = []*core.Route{
	core.NewRedirect("GET", ``, `/static/index.html`),
	core.NewRedirect("GET", `/`, `/static/index.html`),
	core.NewRedirect("GET", `/index.html`, `/static/index.html`),

	core.NewRoute("GET", `/static`, handleStatic),
	core.NewRoute("GET", `/static/.*`, handleStatic),
}

func GetRoutes() *[]*core.Route {
	routes := make([]*core.Route, 0)

	routes = append(routes, baseRoutes...)
	routes = append(routes, *(admin.GetRoutes())...)
	routes = append(routes, *(courses.GetRoutes())...)
	routes = append(routes, *(lms.GetRoutes())...)
	routes = append(routes, *(users.GetRoutes())...)

	return &routes
}
