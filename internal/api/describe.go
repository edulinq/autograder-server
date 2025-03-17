package api

import (
	"errors"
	"reflect"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

// Routes must be validated before calling Describe().
func Describe(routes []core.Route) (*core.APIDescription, error) {
	endpointMap := make(map[string]core.EndpointDescription)
	typeSet := make(map[string]core.TypeDescription)

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

		addType(apiRoute.RequestType, typeSet)
		addType(apiRoute.ResponseType, typeSet)

		endpointMap[apiRoute.GetBasePath()] = core.EndpointDescription{
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
			Description:  apiRoute.Description,
		}
	}

	apiDescription := core.APIDescription{
		Endpoints: endpointMap,
		Types:     typeSet,
	}

	return &apiDescription, errs
}

/* func DescribeTypes(unknownTypes map[string]any) map[string]core.TypeDescription {
    knownTypes := make(map[string]core.TypeDescription)

}*/

func addType(endpointType reflect.Type, typeSet map[string]core.TypeDescription) {
	if endpointType == nil {
		return
	}

	if endpointType.Kind() == reflect.Pointer {
		endpointType = endpointType.Elem()
	}

	typeSet[endpointType.String()] = core.TypeDescription{util.DescribeType(endpointType)}
}
