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

		inputFields := simplifyType(apiRoute.RequestType, typeMap)
		outputFields := simplifyType(apiRoute.ResponseType, typeMap)

		endpointMap[apiRoute.GetBasePath()] = core.EndpointDescription{
			Description:  apiRoute.Description,
			InputFields:  inputFields.StructFields,
			OutputFields: outputFields.StructFields,
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
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

func simplifyType(customType reflect.Type, typeMap map[string]core.TypeDescription) core.TypeDescription {
	if customType == nil {
		return core.TypeDescription{}
	}

	if typeMap == nil {
		typeMap = make(map[string]core.TypeDescription)
	}

	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	typeID := GetTypeID(customType)
	typeDescription, ok := typeMap[typeID]
	if ok {
		return typeDescription
	}

	switch customType.Kind() {
	case reflect.Slice, reflect.Array:
		elementType := customType.Elem()
		elemTypeDescription := simplifyType(elementType, typeMap)

		typeDescription.TypeID = typeID
		typeDescription.TypeCategory = core.ArrayType
		typeDescription.ArrayElementType = elemTypeDescription.TypeID
	case reflect.Map:
		elementType := customType.Elem()
		elemTypeDescription := simplifyType(elementType, typeMap)

		typeDescription.TypeID = typeID
		typeDescription.TypeCategory = core.MapType
		typeDescription.MapKeyType = customType.Key().String()
		typeDescription.MapValueType = elemTypeDescription.TypeID
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
				embeddedFields := simplifyType(field.Type, typeMap)
				if len(embeddedFields.StructFields) > 0 {
					for fieldTag, fieldValue := range embeddedFields.StructFields {
						simplifiedTypes[fieldTag] = fieldValue
					}
				} else {
					simplifiedTypes[jsonTag] = core.TypeDescriptionToString(embeddedFields)
				}

				continue
			}

			// If a field is not embedded, it must have a JSON field name.
			jsonTag = util.JSONFieldNameFull(field, false)
			if jsonTag == "" {
				continue
			}

			fieldDescriptions := simplifyType(field.Type, typeMap)
			simplifiedTypes[jsonTag] = core.TypeDescriptionToString(fieldDescriptions)
		}

		typeDescription.TypeID = typeID
		typeDescription.TypeCategory = core.StructType
		if len(simplifiedTypes) > 0 {
			typeDescription.StructFields = simplifiedTypes
		}
	default:
		// Handle built-in types.
		typeDescription.TypeID = typeID
		typeDescription.TypeCategory = core.BasicType

		if customType.PkgPath() == "" {
			typeDescription.Alias = customType.String()
		} else {
			typeDescription.Alias = customType.Kind().String()
		}
	}

	if typeDescription.TypeCategory != core.BasicType {
		typeMap[typeID] = typeDescription
	}

	return typeDescription
}
