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
	core.NewRedirect("GET", ``, `/static/index.html`),
	core.NewRedirect("GET", `/`, `/static/index.html`),
	core.NewRedirect("GET", `/index.html`, `/static/index.html`),

	core.NewBaseRoute("GET", `/static`, static.Handle),
	core.NewBaseRoute("GET", `/static/.*`, static.Handle),
}

type APIDescription struct {
	Endpoints map[string]map[string]EndpointDescription
}

type EndpointDescription struct {
	Description  string
	RequestType  string
	ResponseType string
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
		if ok {
			endpointMap[apiRoute.GetBasePath()] = EndpointDescription{
				Description:  apiRoute.GetDescription(),
				RequestType:  apiRoute.Request.String(),
				ResponseType: apiRoute.Response.String(),
			}
		}
	}

	return &APIDescription{
		Endpoints: map[string]map[string]EndpointDescription{
			"endpoints": endpointMap,
		},
	}
}

func DescribeToJSON(description *APIDescription) (string, error) {
	if description == nil {
		return "", nil
	}

	return util.ToJSONIndent(description.Endpoints)
}
