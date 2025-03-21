package api

import (
	"errors"
	"reflect"
	"strings"

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

		inputFields, _ := simplifyType(apiRoute.RequestType, typeMap)
		outputFields, _ := simplifyType(apiRoute.ResponseType, typeMap)

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

func simplifyType(customType reflect.Type, typeMap map[string]core.TypeDescription) (map[string]string, string) {
	if customType == nil {
		return map[string]string{}, "<nil>"
	}

	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	typeID := getTypeID(customType)
	typeDescription, ok := typeMap[typeID]
	if ok {
		return typeDescription, ""
	}

	simplifiedTypes := make(map[string]string)

	switch customType.Kind() {
	case reflect.Slice, reflect.Array:
		return simplifiedTypes, simplifyArrayType(customType, typeMap)
	case reflect.Map:
		return simplifiedTypes, simplifyMapType(customType, typeMap)
	case reflect.Struct:
		for i := 0; i < customType.NumField(); i++ {
			field := customType.Field(i)
			jsonTag := util.JSONFieldName(field)
			if jsonTag == "" {
				continue
			}

			fieldDescriptions, simpleDescription := simplifyType(field.Type, typeMap)
			if len(fieldDescriptions) > 0 {
				simplifiedTypes[jsonTag] = formatNestedFields(fieldDescriptions)
			} else {
				simplifiedTypes[jsonTag] = simpleDescription
			}
		}
	default:
		// Handle built-in types.
		if customType.PkgPath() == "" {
			return simplifiedTypes, customType.String()
		} else {
			return simplifiedTypes, customType.Kind().String()
		}
	}

	typeMap[typeID] = simplifiedTypes
	return simplifiedTypes, ""
}

func simplifyArrayType(customType reflect.Type, typeMap map[string]core.TypeDescription) string {
	if customType == nil {
		return "<nil>"
	}

	if customType.Kind() != reflect.Slice && customType.Kind() != reflect.Array {
		return "<error: not an array type>"
	}

	elementType := customType.Elem()
	elementDescription, simpleDescription := simplifyType(elementType, typeMap)

	var arrayDescription string
	if len(elementDescription) > 0 {
		arrayDescription = "[]" + formatNestedFields(elementDescription)
	} else {
		arrayDescription = "[]" + simpleDescription
	}

	return arrayDescription
}

func simplifyMapType(customType reflect.Type, typeMap map[string]core.TypeDescription) string {
	if customType == nil {
		return "<nil>"
	}

	if customType.Kind() != reflect.Map {
		return "<error: not a map type>"
	}

	keyType := customType.Key()
	elementType := customType.Elem()
	elementDescription, simpleDescription := simplifyType(elementType, typeMap)

	var mapDescription string
	if len(elementDescription) > 0 {
		mapDescription = "map[" + keyType.String() + "]{" + formatNestedFields(elementDescription) + "}"
	} else {
		mapDescription = "map[" + keyType.String() + "]{" + simpleDescription + "}"
	}

	return mapDescription
}

func getTypeID(customType reflect.Type) string {
	return customType.PkgPath() + "/" + customType.String()
}

func formatNestedFields(fields map[string]string) string {
	parts := []string{}
	for key, value := range fields {
		parts = append(parts, key+": "+value)
	}

	return strings.Join(parts, ", ")
}
