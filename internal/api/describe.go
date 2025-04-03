package api

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var skipDescriptionPatterns = []*regexp.Regexp{
	regexp.MustCompile("root-user-nonce"),
	regexp.MustCompile("Min.*Role.*"),
	regexp.MustCompile("^APIRequest$"),
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

		// RequestType and ResponseType must be structs, so Fields will hold the type's information.
		input, _, typeMap, _, err := describeType(apiRoute.RequestType, false, typeMap, nil)
		if err != nil {
			errs = errors.Join(errs, err)
		}

		output, _, typeMap, _, err := describeType(apiRoute.ResponseType, false, typeMap, nil)
		if err != nil {
			errs = errors.Join(errs, err)
		}

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

func getTypeID(customType reflect.Type, typeConversions map[string]string) (string, error) {
	if typeConversions == nil {
		typeConversions = make(map[string]string)
	}

	prefix := ""

	// Include the PkgPath() of pointers to custom types.
	if customType.Kind() == reflect.Pointer {
		prefix = "*"
		customType = customType.Elem()
	}

	if customType.PkgPath() != "" {
		longTypeName := filepath.Dir(customType.PkgPath()) + "/" + customType.String()
		// Check if a type simplification exists for the short type name.
		shortTypeName, ok := typeConversions[longTypeName]
		if ok {
			return prefix + shortTypeName, nil
		}

		// If a type simplification does not exist, we must create one.
		shortTypeName = customType.String()
		competingLongName, ok := typeConversions[shortTypeName]
		if !ok {
			typeConversions[shortTypeName] = longTypeName
			typeConversions[longTypeName] = shortTypeName
			return prefix + shortTypeName, nil
		}

		return "", fmt.Errorf("Unable to get type ID due to conflicting names. Both '%s' and '%s' share the ID: '%s'.",
			competingLongName, longTypeName, shortTypeName)
	} else {
		if customType.Kind() == reflect.Array || customType.Kind() == reflect.Slice {
			elemTypeID, err := getTypeID(customType.Elem(), typeConversions)
			if err != nil {
				return "", err
			}

			return prefix + "[]" + elemTypeID, nil
		}

		if customType.Kind() == reflect.Map {
			keyTypeID, err := getTypeID(customType.Key(), typeConversions)
			if err != nil {
				return "", err
			}

			valueTypeID, err := getTypeID(customType.Elem(), typeConversions)
			if err != nil {
				return "", err
			}

			return prefix + "map[" + keyTypeID + "]" + valueTypeID, nil
		}
	}

	return prefix + customType.String(), nil
}

func mustGetTypeID(customType reflect.Type, typeConversions map[string]string) string {
	typeID, err := getTypeID(customType, typeConversions)
	if err != nil {
		log.Fatal("Failed to get type ID.", err)
	}

	return typeID
}

// Given a type and a map of known type descriptions, describeType() returns the type description and typeID.
//   - Basic types (PODs) return their String() as both an Alias and typeID.
//   - Arrays and slices store the typeID of their element in ElementType.
//   - Maps store the key type as a string and the value as a typeID in KeyType and ValueType respectively.
//   - Structs have a Fields map describing each field, including embedded ones.
//     Non-embedded struct fields that do not have a JSON tag are skipped.
func describeType(customType reflect.Type, addType bool, typeMap map[string]core.TypeDescription, typeConversions map[string]string) (core.TypeDescription, string, map[string]core.TypeDescription, map[string]string, error) {
	if customType == nil {
		return core.TypeDescription{}, "", map[string]core.TypeDescription{}, map[string]string{}, fmt.Errorf("Unable to describe nil type.")
	}

	if typeMap == nil {
		typeMap = make(map[string]core.TypeDescription)
	}

	if typeConversions == nil {
		typeConversions = make(map[string]string)
	}

	originalTypeID, err := getTypeID(customType, typeConversions)
	if err != nil {
		return core.TypeDescription{}, "", map[string]core.TypeDescription{}, map[string]string{}, err
	}

	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	typeID, err := getTypeID(customType, typeConversions)
	if err != nil {
		return core.TypeDescription{}, "", map[string]core.TypeDescription{}, map[string]string{}, err
	}

	typeDescription, ok := typeMap[typeID]
	if ok {
		return typeDescription, originalTypeID, typeMap, typeConversions, nil
	}

	switch customType.Kind() {
	case reflect.Slice, reflect.Array:
		_, elemTypeID, _, _, err := describeType(customType.Elem(), true, typeMap, typeConversions)
		if err != nil {
			return core.TypeDescription{}, "", map[string]core.TypeDescription{}, map[string]string{}, err
		}

		typeDescription.Category = core.ArrayType
		typeDescription.ElementType = elemTypeID
	case reflect.Map:
		_, elemTypeID, _, _, err := describeType(customType.Elem(), true, typeMap, typeConversions)
		if err != nil {
			return core.TypeDescription{}, "", map[string]core.TypeDescription{}, map[string]string{}, err
		}

		_, keyTypeID, _, _, err := describeType(customType.Key(), true, typeMap, typeConversions)
		if err != nil {
			return core.TypeDescription{}, "", map[string]core.TypeDescription{}, map[string]string{}, err
		}

		typeDescription.Category = core.MapType
		typeDescription.KeyType = keyTypeID
		typeDescription.ValueType = elemTypeID
	case reflect.Struct:
		typeDescription.Category = core.StructType
		typeDescription.Fields, err = describeStructFields(customType, typeMap, typeConversions)
		if err != nil {
			return core.TypeDescription{}, "", map[string]core.TypeDescription{}, map[string]string{}, err
		}
	default:
		// Handle built-in types.
		typeDescription.Category = core.AliasType

		if customType.PkgPath() == "" {
			typeDescription.AliasType = customType.String()
		} else {
			typeDescription.AliasType = customType.Kind().String()
		}
	}

	if addType && customType.PkgPath() != "" {
		typeMap[typeID] = typeDescription
	}

	return typeDescription, originalTypeID, typeMap, typeConversions, nil
}

func describeStructFields(customType reflect.Type, typeMap map[string]core.TypeDescription, typeConversions map[string]string) (map[string]string, error) {
	fieldTypes := make(map[string]string)

	for i := 0; i < customType.NumField(); i++ {
		field := customType.Field(i)

		jsonTag := util.JSONFieldName(field)
		if jsonTag == "" {
			continue
		}

		if skipField(jsonTag) {
			continue
		}

		// Handle embedded fields.
		if field.Anonymous {
			fieldDescription, fieldTypeID, _, _, err := describeType(field.Type, true, typeMap, typeConversions)
			if err != nil {
				return map[string]string{}, err
			}

			// If the embedded type is a struct, merge its fields into the current struct.
			if len(fieldDescription.Fields) > 0 {
				for fieldTag, fieldValue := range fieldDescription.Fields {
					if skipField(fieldTag) {
						continue
					}

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

		fieldDescription, fieldTypeID, _, _, err := describeType(field.Type, true, typeMap, typeConversions)
		if err != nil {
			return map[string]string{}, err
		}

		if fieldDescription.Category == core.AliasType {
			fieldTypes[jsonTag] = fieldDescription.AliasType
		} else {
			fieldTypes[jsonTag] = fieldTypeID
		}
	}

	return fieldTypes, nil
}

func skipField(name string) bool {
	for _, pattern := range skipDescriptionPatterns {
		if pattern.MatchString(name) {
			return true
		}
	}

	return false
}
