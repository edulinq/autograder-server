package api

import (
	"github.com/edulinq/autograder/internal/api/core"
)

// Routes must be validated before calling Describe.
func Describe(routes []core.Route) *core.APIDescription {
	endpointMap := make(map[string]core.EndpointDescription)
	for _, route := range routes {
		apiRoute, ok := route.(*core.APIRoute)
		if !ok {
			continue
		}

		endpointMap[apiRoute.GetBasePath()] = core.EndpointDescription{
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
		}
	}

	return &core.APIDescription{
		Endpoints: endpointMap,
	}
}
