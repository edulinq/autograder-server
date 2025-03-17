package api

import (
	"errors"
	"reflect"

	"github.com/edulinq/autograder/internal/api/core"
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

func addType(endpointType reflect.Type, typeSet map[string]core.TypeDescription) {
	if endpointType == nil {
		return
	}

	if endpointType.Kind() == reflect.Pointer {
		endpointType = endpointType.Elem()
	}

	// Skip built in types.
	if endpointType.PkgPath() == "" {
		return
	}

	typeName := endpointType.String()
	_, ok := typeSet[typeName]
	if ok {
		return
	}

	if endpointType.Kind() != reflect.Struct {
		typeSet[typeName] = core.TypeDescription{
			Alias: endpointType.Kind().String(),
		}
		return
	}

	typeSet[typeName] = core.TypeDescription{
		Fields: describeType(endpointType, typeSet),
	}
}

func describeType(reflectType reflect.Type, typeSet map[string]core.TypeDescription) map[string]string {
	if reflectType == nil {
		return map[string]string{}
	}

	if reflectType.Kind() == reflect.Pointer {
		reflectType = reflectType.Elem()
	}

	if reflectType.Kind() != reflect.Struct {
		return map[string]string{}
	}

	description := make(map[string]string)

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		fieldType := field.Type

		if fieldType.Kind() == reflect.Pointer || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
			fieldType = fieldType.Elem()
		}

		description[field.Name] = field.Type.String()

		addType(fieldType, typeSet)
	}

	return description
}
