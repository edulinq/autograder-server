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

		// RequestType and ResponseType must be structs, so Fields will hold the simplified types.
		simplifiedRequest, _ := simplifyType(apiRoute.RequestType, typeMap)
		simplifiedResponse, _ := simplifyType(apiRoute.ResponseType, typeMap)

		endpointMap[apiRoute.GetBasePath()] = core.EndpointDescription{
			Description:  apiRoute.Description,
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
			InputFields:  simplifiedRequest.Fields,
			OutputFields: simplifiedResponse.Fields,
		}
	}

	apiDescription := core.APIDescription{
		Endpoints: endpointMap,
		Types:     typeMap,
	}

	return &apiDescription, errs
}

func GetTypeID(customType reflect.Type) string {
	typeID := customType.String()
	if customType.PkgPath() != "" {
		typeID = customType.PkgPath() + "/" + typeID
	}

	return typeID
}

func simplifyType(customType reflect.Type, typeMap map[string]core.TypeDescription) (core.TypeDescription, string) {
	if customType == nil {
		return core.TypeDescription{}, ""
	}

	// If typeMap is nil, the results will not be accessible outside of the function.
	if typeMap == nil {
		typeMap = make(map[string]core.TypeDescription)
	}

	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	typeID := GetTypeID(customType)
	typeDescription, ok := typeMap[typeID]
	if ok {
		return typeDescription, typeID
	}

	switch customType.Kind() {
	case reflect.Slice, reflect.Array:
		_, elemTypeID := simplifyType(customType.Elem(), typeMap)

		typeDescription.Category = core.ArrayType
		typeDescription.ElementType = elemTypeID
	case reflect.Map:
		_, elemTypeID := simplifyType(customType.Elem(), typeMap)

		typeDescription.Category = core.MapType
		typeDescription.KeyType = customType.Key().String()
		typeDescription.ValueType = elemTypeID
	case reflect.Struct:
		simplifiedTypes := make(map[string]string)

		for i := 0; i < customType.NumField(); i++ {
			field := customType.Field(i)

			jsonTag := util.JSONFieldName(field)
			if jsonTag == "" {
				continue
			}

			// Handle embedded fields.
			if field.Anonymous {
				fieldDescription, fieldTypeID := simplifyType(field.Type, typeMap)
				if len(fieldDescription.Fields) > 0 {
					for fieldTag, fieldValue := range fieldDescription.Fields {
						simplifiedTypes[fieldTag] = fieldValue
					}
				} else if fieldDescription.Category == core.BasicType {
					simplifiedTypes[jsonTag] = fieldDescription.Alias
				} else {
					simplifiedTypes[jsonTag] = fieldTypeID
				}

				continue
			}

			// Non-embedded fields must have a JSON field name.
			jsonTag = util.JSONFieldNameFull(field, false)
			if jsonTag == "" {
				continue
			}

			fieldDescription, fieldTypeID := simplifyType(field.Type, typeMap)
			if fieldDescription.Category == core.BasicType {
				simplifiedTypes[jsonTag] = fieldDescription.Alias
			} else {
				simplifiedTypes[jsonTag] = fieldTypeID
			}
		}

		typeDescription.Category = core.StructType
		if len(simplifiedTypes) > 0 {
			typeDescription.Fields = simplifiedTypes
		}
	default:
		// Handle built-in types.
		typeDescription.Category = core.BasicType

		if customType.PkgPath() == "" {
			typeDescription.Alias = customType.String()
		} else {
			typeDescription.Alias = customType.Kind().String()
		}
	}

	if typeDescription.Category != core.BasicType {
		typeMap[typeID] = typeDescription
	}

	return typeDescription, typeID
}
