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
	typeMap := make(map[string]core.TypeDescription)

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

		// TODO: Change to discoverType?
		describeType(apiRoute.RequestType, typeMap)
		describeType(apiRoute.ResponseType, typeMap)

		inputFields := resolveType(apiRoute.RequestType, typeMap)
		outputFields := resolveType(apiRoute.ResponseType, typeMap)

		endpointMap[apiRoute.GetBasePath()] = core.EndpointDescription{
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
			Description:  apiRoute.Description,
			InputFields:  inputFields,
			OutputFields: outputFields,
		}
	}

	apiDescription := core.APIDescription{
		Endpoints: endpointMap,
	}

	return &apiDescription, errs
}

func describeType(customType reflect.Type, typeMap map[string]core.TypeDescription) {
	if customType == nil {
		return
	}

	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	// Skip built-in types.
	if customType.PkgPath() == "" {
		return
	}

	typeID := getTypeID(customType)
	// TODO: May add PkgPath to avoid type conflicts.
	// We won't show this to the user so it's okay.
	// Can we use the type itself as the key?
	// If not, can we use the pointer to the type object?
	// typeName := customType.String()
	_, ok := typeMap[typeID]
	if ok {
		return
	}

	// Custom types that are not structs are an alias for another type.
	if customType.Kind() != reflect.Struct {
		typeMap[typeID] = core.TypeDescription{
			Alias: customType.Kind().String(),
		}
		return
	}

	fieldDescriptions := make(map[string]string)

	for i := 0; i < customType.NumField(); i++ {
		field := customType.Field(i)
		fieldType := field.Type

		if fieldType.Kind() == reflect.Pointer || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
			fieldType = fieldType.Elem()
		}

		fieldDescriptions[field.Name] = field.Type.String()

		describeType(fieldType, typeMap)
	}

	typeMap[typeID] = core.TypeDescription{
		Fields: fieldDescriptions,
	}
}

func resolveType(customType reflect.Type, typeMap map[string]core.TypeDescription) map[string]string {
	resolvedFields := make(map[string]string)

	if customType == nil {
		return resolvedFields
	}

	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	// If it's a built-in type, return it directly.
	if customType.PkgPath() == "" {
		return map[string]string{"value": customType.String()}
	}

	if customType.Kind() == reflect.Struct {
		for i := 0; i < customType.NumField(); i++ {
			field := customType.Field(i)

			jsonTag := util.JSONFieldName(field)
			if jsonTag == "" {
				continue
			}

			fieldType := field.Type

			for fieldType.Kind() == reflect.Pointer || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
				fieldType = fieldType.Elem()
			}

			baseType := fieldType.String()
			typeID := getTypeID(fieldType)

			// Resolve complex types by looking through the type map.
			typeDescription, ok := typeMap[typeID]
			if ok {
				if typeDescription.Alias != "" {
					baseType = typeDescription.Alias
				} else if len(typeDescription.Fields) > 0 {
					nestedFields := resolveType(fieldType, typeMap)
					for nestedTag, nestedType := range nestedFields {
						resolvedFields[nestedTag] = nestedType
					}
					continue
				}
			}

			resolvedFields[jsonTag] = baseType
		}
	}

	return resolvedFields
}

func getTypeID(customType reflect.Type) string {
	return customType.PkgPath() + "/" + customType.String()
}
