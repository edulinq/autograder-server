package api

// All routes handled by the server.

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses"
	"github.com/edulinq/autograder/internal/api/lms"
	"github.com/edulinq/autograder/internal/api/logs"
	"github.com/edulinq/autograder/internal/api/static"
	"github.com/edulinq/autograder/internal/api/users"
)

var baseRoutes = []core.Route{
	core.NewRedirect("GET", ``, `/static/index.html`),
	core.NewRedirect("GET", `/`, `/static/index.html`),
	core.NewRedirect("GET", `/index.html`, `/static/index.html`),

	core.NewBaseRoute("GET", `/static`, static.Handle),
	core.NewBaseRoute("GET", `/static/.*`, static.Handle),
}

type APIDescription struct {
	Endpoints map[string]EndpointDescription `json:"endpoints"`
}

type EndpointDescription struct {
	RequestType  string `json:"request-type"`
	ResponseType string `json:"response-type"`
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

func Describe() *APIDescription {
	routes := GetRoutes()

	endpointMap := make(map[string]EndpointDescription)
	for _, route := range *routes {
		apiRoute, ok := route.(*core.APIRoute)
		if !ok {
			continue
		}

		endpointMap[apiRoute.GetBasePath()] = EndpointDescription{
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
		}
	}

	return &APIDescription{
		Endpoints: endpointMap,
	}
}
