package api

import (
	"errors"

	"github.com/edulinq/autograder/internal/api/core"
)

// Routes must be validated before calling Describe().
func Describe(routes []core.Route) (*core.APIDescription, error) {
	endpointMap := make(map[string]core.EndpointDescription)
	var errs error = nil
	var err error
	for _, route := range routes {
		apiRoute, ok := route.(*core.APIRoute)
		if !ok {
			continue
		}

		// Check if we have already found the description for an endpoint.
		if apiRoute.Description == "" {
			apiRoute.Description, err = core.GetDescriptionFromHandler(apiRoute.GetBasePath())
			if err != nil {
				errs = errors.Join(errs, err)
			}
		}

		endpointMap[apiRoute.GetBasePath()] = core.EndpointDescription{
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
			Description:  apiRoute.Description,
		}
	}

	apiDescription := core.APIDescription{
		Endpoints: endpointMap,
	}

	return &apiDescription, errs
}
