package api

// All routes handled by the server.

import (
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses"
	"github.com/edulinq/autograder/internal/api/lms"
	"github.com/edulinq/autograder/internal/api/logs"
	"github.com/edulinq/autograder/internal/api/static"
	"github.com/edulinq/autograder/internal/api/users"
)

var baseRoutes = []core.Route{
	core.NewRedirect("GET", ``, `/static/index.html`, "TODO: Description."),
	core.NewRedirect("GET", `/`, `/static/index.html`, "TODO: Description."),
	core.NewRedirect("GET", `/index.html`, `/static/index.html`, "TODO: Description."),

	core.NewRoute("GET", `/static`, static.Handle, "TODO: Description."),
	core.NewRoute("GET", `/static/.*`, static.Handle, "TODO: Description."),
}

func GetRoutes() *[]core.Route {
	routes := make([]core.Route, 0)

	routes = append(routes, baseRoutes...)
	routes = append(routes, *(courses.GetRoutes())...)
	routes = append(routes, *(lms.GetRoutes())...)
	routes = append(routes, *(logs.GetRoutes())...)
	routes = append(routes, *(users.GetRoutes())...)

	return &routes
}

func Describe() string {
	var builder strings.Builder

	routes := GetRoutes()
	for _, route := range *routes {
		builder.WriteString(route.GetSuffix())
		builder.WriteString("\n")
	}

	return builder.String()
}
