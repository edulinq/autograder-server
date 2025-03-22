package api

import (
	"errors"
	"reflect"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

type internalTypeDescription struct {
	Description core.TypeDescription
	Alias       string
}

func (this *internalTypeDescription) String() string {
	if len(this.Description) > 0 {
		return formatNestedFields(this.Description)
	} else {
		return this.Alias
	}
}

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
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
			Description:  apiRoute.Description,
			InputFields:  inputFields.Description,
			OutputFields: outputFields.Description,
		}
	}

	apiDescription := core.APIDescription{
		Endpoints: endpointMap,
	}

	return &apiDescription, errs
}

func simplifyType(customType reflect.Type, typeMap map[string]core.TypeDescription) internalTypeDescription {
	if customType == nil {
		return internalTypeDescription{Alias: "<nil>"}
	}

	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	typeID := getTypeID(customType)
	typeDescription, ok := typeMap[typeID]
	if ok {
		return internalTypeDescription{Description: typeDescription}
	}

	simplifiedTypes := make(map[string]string)

	switch customType.Kind() {
	case reflect.Slice, reflect.Array:
		return simplifyArrayType(customType, typeMap)
	case reflect.Map:
		return simplifyMapType(customType, typeMap)
	case reflect.Struct:
		for i := 0; i < customType.NumField(); i++ {
			field := customType.Field(i)

			jsonTag := util.JSONFieldName(field)
			if jsonTag == "" {
				continue
			}

			// Handle embedded fields.
			if field.Anonymous {
				embeddedFields := simplifyType(field.Type, typeMap)
				if len(embeddedFields.Description) > 0 {
					for fieldTag, fieldValue := range embeddedFields.Description {
						simplifiedTypes[fieldTag] = fieldValue
					}
				} else if embeddedFields.Alias != "" {
					simplifiedTypes[jsonTag] = embeddedFields.Alias
				}

				continue
			}

			fieldDescriptions := simplifyType(field.Type, typeMap)
			simplifiedTypes[jsonTag] = fieldDescriptions.String()
		}
	default:
		// Handle built-in types.
		if customType.PkgPath() == "" {
			return internalTypeDescription{Alias: customType.String()}
		} else {
			return internalTypeDescription{Alias: customType.Kind().String()}
		}
	}

	typeMap[typeID] = simplifiedTypes
	return internalTypeDescription{Description: simplifiedTypes}
}

func simplifyArrayType(customType reflect.Type, typeMap map[string]core.TypeDescription) internalTypeDescription {
	if customType == nil {
		return internalTypeDescription{Alias: "<nil>"}
	}

	if customType.Kind() != reflect.Slice && customType.Kind() != reflect.Array {
		return internalTypeDescription{Alias: "<error: not an array type>"}
	}

	elementType := customType.Elem()
	typeDescription := simplifyType(elementType, typeMap)

	arrayDescription := "[]{" + typeDescription.String() + "}"

	return internalTypeDescription{Alias: arrayDescription}
}

func simplifyMapType(customType reflect.Type, typeMap map[string]core.TypeDescription) internalTypeDescription {
	if customType == nil {
		return internalTypeDescription{Alias: "<nil>"}
	}

	if customType.Kind() != reflect.Map {
		return internalTypeDescription{Alias: "<error: not a map type>"}
	}

	keyType := customType.Key()
	elementType := customType.Elem()
	typeDescription := simplifyType(elementType, typeMap)

	mapDescription := "map[" + keyType.String() + "]{" + typeDescription.String() + "}"

	return internalTypeDescription{Alias: mapDescription}
}

func getTypeID(customType reflect.Type) string {
	return customType.PkgPath() + "/" + customType.String()
}

func formatNestedFields(fields map[string]string) string {
	parts := []string{}

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	for _, key := range keys {
		parts = append(parts, key+": "+fields[key])
	}

	return strings.Join(parts, ", ")
}
