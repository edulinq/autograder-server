package api

// All routes handled by the server.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses"
	"github.com/edulinq/autograder/internal/api/lms"
	"github.com/edulinq/autograder/internal/api/logs"
	"github.com/edulinq/autograder/internal/api/static"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/util"
)

var baseRoutes = []core.Route{
	core.NewRedirect("GET", ``, `/static/index.html`, "Redirects to the main static index page."),
	core.NewRedirect("GET", `/`, `/static/index.html`, "Redirects the root path to the main static index page."),
	core.NewRedirect("GET", `/index.html`, `/static/index.html`, "Redirects the index page to the main static index page."),

	core.NewRoute("GET", `/static`, static.Handle, "Serves static resources."),
	core.NewRoute("GET", `/static/.*`, static.Handle, "Serves static resources for any path under '/static'."),
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

func Describe() (string, error) {
	routes := GetRoutes()

	endpointMap := make(map[string]map[string]string)
	for _, route := range *routes {
		apiRoute, ok := route.(*core.APIRoute)
		if ok {
			endpointMap[apiRoute.GetSuffix()] = map[string]string{
				"description":  apiRoute.GetDescription(),
				"requestType":  apiRoute.Request.String(),
				"responseType": apiRoute.Response.String(),
			}
		}
	}

	endpoints := map[string]any{
		"endpoints": endpointMap,
	}

	return util.ToJSONIndent(endpoints)
}
