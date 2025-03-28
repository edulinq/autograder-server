package api

import (
	"errors"
	"path/filepath"
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
		input, _, typeMap := describeType(apiRoute.RequestType, typeMap)
		output, _, typeMap := describeType(apiRoute.ResponseType, typeMap)

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

func getTypeID(customType reflect.Type) string {
	prefix := ""

	// Include the PkgPath() of pointers to custom types.
	if customType.Kind() == reflect.Pointer {
		prefix = "*"
		customType = customType.Elem()
	}

	if customType.PkgPath() != "" {
		prefix = prefix + filepath.Dir(customType.PkgPath()) + "/"
	} else {
		if customType.Kind() == reflect.Array || customType.Kind() == reflect.Slice {
			return prefix + "[]" + getTypeID(customType.Elem())
		}

		if customType.Kind() == reflect.Map {
			return prefix + "map[" + getTypeID(customType.Key()) + "]" + getTypeID(customType.Elem())
		}
	}

	return prefix + customType.String()
}

// Given a type and a map of known type descriptions, describeType() returns the type description and typeID.
//   - Basic types (PODs) return their String() as both an Alias and typeID.
//   - Arrays and slices store the typeID of their element in ElementType.
//   - Maps store the key type as a string and the value as a typeID in KeyType and ValueType respectively.
//   - Structs have a Fields map describing each field, including embedded ones.
//     Non-embedded struct fields that do not have a JSON tag are skipped.
func describeType(customType reflect.Type, typeMap map[string]core.TypeDescription) (core.TypeDescription, string, map[string]core.TypeDescription) {
	if customType == nil {
		return core.TypeDescription{}, "", map[string]core.TypeDescription{}
	}

	if typeMap == nil {
		typeMap = make(map[string]core.TypeDescription)
	}

	originalTypeID := getTypeID(customType)
	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	typeID := getTypeID(customType)
	typeDescription, ok := typeMap[typeID]
	if ok {
		return typeDescription, originalTypeID, typeMap
	}

	switch customType.Kind() {
	case reflect.Slice, reflect.Array:
		_, elemTypeID, _ := describeType(customType.Elem(), typeMap)

		typeDescription.Category = core.ArrayType
		typeDescription.ElementType = elemTypeID
	case reflect.Map:
		_, elemTypeID, _ := describeType(customType.Elem(), typeMap)

		typeDescription.Category = core.MapType
		typeDescription.KeyType = customType.Key().String()
		typeDescription.ValueType = elemTypeID
	case reflect.Struct:
		typeDescription.Category = core.StructType
		typeDescription.Fields = describeStructFields(customType, typeMap)
	default:
		// Handle built-in types.
		typeDescription.Category = core.AliasType

		if customType.PkgPath() == "" {
			typeDescription.AliasType = customType.String()
		} else {
			typeDescription.AliasType = customType.Kind().String()
		}
	}

	if customType.PkgPath() != "" {
		typeMap[typeID] = typeDescription
	}

	return typeDescription, originalTypeID, typeMap
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
			fieldDescription, fieldTypeID, _ := describeType(field.Type, typeMap)
			// If the embedded type is a struct, merge its fields into the current struct.
			if len(fieldDescription.Fields) > 0 {
				for fieldTag, fieldValue := range fieldDescription.Fields {
					fieldTypes[fieldTag] = fieldValue
				}
			} else if fieldDescription.Category == core.AliasType {
				// Store basic embedded types under the current JSON tag.
				fieldTypes[jsonTag] = fieldDescription.AliasType
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

		fieldDescription, fieldTypeID, _ := describeType(field.Type, typeMap)
		if fieldDescription.Category == core.AliasType {
			fieldTypes[jsonTag] = fieldDescription.AliasType
		} else {
			fieldTypes[jsonTag] = fieldTypeID
		}
	}

	return fieldTypes
}
