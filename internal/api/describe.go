package api

import (
	"github.com/edulinq/autograder/internal/api/core"
)

func Describe(routes []core.Route) *core.APIDescription {
	endpointMap := make(map[string]core.EndpointDescription)
	for _, route := range routes {
		apiRoute, ok := route.(*core.APIRoute)
		if !ok {
			continue
		}

		var requestTypeStr, responseTypeStr string

		if apiRoute.RequestType != nil {
			requestTypeStr = apiRoute.RequestType.String()
		} else {
			requestTypeStr = "<nil>"
		}

		if apiRoute.ResponseType != nil {
			responseTypeStr = apiRoute.ResponseType.String()
		} else {
			responseTypeStr = "<nil>"
		}

		endpointMap[apiRoute.GetBasePath()] = core.EndpointDescription{
			RequestType:  requestTypeStr,
			ResponseType: responseTypeStr,
		}
	}

	return &core.APIDescription{
		Endpoints: endpointMap,
	}
}
