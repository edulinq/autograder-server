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

		// RequestType and ResponseType must be structs, so Fields will hold the type's information.
		input, _ := describeType(apiRoute.RequestType, typeMap)
		output, _ := describeType(apiRoute.ResponseType, typeMap)

		endpointMap[apiRoute.GetBasePath()] = core.EndpointDescription{
			Description:  apiRoute.Description,
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
			Input:        input.Fields,
			Output:       output.Fields,
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

	// Include the PkgPath() of pointers to custom types.
	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	if customType.PkgPath() != "" {
		typeID = customType.PkgPath() + "/" + typeID
	}

	return typeID
}

// Given a type and a map of known type descriptions, describeType() returns the type description and typeID.
//   - Basic types (PODs) return their String() as both an Alias and typeID.
//   - Arrays and slices store the typeID of their element in ElementType.
//   - Maps store the key type as a string and the value as a typeID in KeyType and ValueType respectively.
//   - Structs have a Fields map describing each field, including embedded ones.
//     Non-embedded struct fields that do not have a JSON tag are skipped.
func describeType(customType reflect.Type, typeMap map[string]core.TypeDescription) (core.TypeDescription, string) {
	if customType == nil {
		return core.TypeDescription{}, ""
	}

	// If typeMap is nil, the results will not be accessible outside of the function.
	if typeMap == nil {
		typeMap = make(map[string]core.TypeDescription)
	}

	originalTypeID := GetTypeID(customType)
	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	typeID := GetTypeID(customType)
	typeDescription, ok := typeMap[typeID]
	if ok {
		return typeDescription, originalTypeID
	}

	switch customType.Kind() {
	case reflect.Slice, reflect.Array:
		_, elemTypeID := describeType(customType.Elem(), typeMap)

		typeDescription.Category = core.ArrayType
		typeDescription.ElementType = elemTypeID
	case reflect.Map:
		_, elemTypeID := describeType(customType.Elem(), typeMap)

		typeDescription.Category = core.MapType
		typeDescription.KeyType = customType.Key().String()
		typeDescription.ValueType = elemTypeID
	case reflect.Struct:
		typeDescription.Category = core.StructType
		typeDescription.Fields = describeStructFields(customType, typeMap)
	default:
		// Handle built-in types.
		typeDescription.Category = core.BasicType

		if customType.PkgPath() == "" {
			typeDescription.Alias = customType.String()
		} else {
			typeDescription.Alias = customType.Kind().String()
		}
	}

	if customType.PkgPath() != "" {
		typeMap[typeID] = typeDescription
	}

	return typeDescription, originalTypeID
}

func describeStructFields(customType reflect.Type, typeMap map[string]core.TypeDescription) map[string]string {
	fieldTypes := make(map[string]string)

	for i := 0; i < customType.NumField(); i++ {
		field := customType.Field(i)

		jsonTag := util.JSONFieldName(field)
		if jsonTag == "" {
			continue
		}

		// Handle embedded fields.
		if field.Anonymous {
			fieldDescription, fieldTypeID := describeType(field.Type, typeMap)
			// If the embedded type is a struct, merge its fields into the current struct.
			if len(fieldDescription.Fields) > 0 {
				for fieldTag, fieldValue := range fieldDescription.Fields {
					fieldTypes[fieldTag] = fieldValue
				}
			} else if fieldDescription.Category == core.BasicType {
				// Store basic embedded types under the current JSON tag.
				fieldTypes[jsonTag] = fieldDescription.Alias
			} else {
				// Store non-basic embedded types under the current JSON tag using the typeID.
				fieldTypes[jsonTag] = fieldTypeID
			}

			continue
		}

		// Non-embedded fields must have a JSON field name.
		jsonTag = util.JSONFieldNameFull(field, false)
		if jsonTag == "" {
			continue
		}

		fieldDescription, fieldTypeID := describeType(field.Type, typeMap)
		if fieldDescription.Category == core.BasicType {
			fieldTypes[jsonTag] = fieldDescription.Alias
		} else {
			fieldTypes[jsonTag] = fieldTypeID
		}
	}

	if len(fieldTypes) > 0 {
		return fieldTypes
	} else {
		return nil
	}
}
