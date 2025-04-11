package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

const (
	AliasType  = "alias"
	ArrayType  = "array"
	MapType    = "map"
	StructType = "struct"
)

var skipDescriptionPatterns = []*regexp.Regexp{
	regexp.MustCompile("^root-user-nonce$"),
	regexp.MustCompile("^Min.*Role.*$"),
	regexp.MustCompile("^APIRequest$"),
}

// API Description will be empty until RunServer() is called.
var apiDescription *APIDescription = nil

type APIDescription struct {
	Endpoints map[string]EndpointDescription `json:"endpoints"`
	Types     map[string]TypeDescription     `json:"types"`
}

type EndpointDescription struct {
	Description  string             `json:"description"`
	RequestType  string             `json:"-"`
	ResponseType string             `json:"-"`
	Input        []FieldDescription `json:"input"`
	Output       []FieldDescription `json:"output"`
}

type FieldDescription struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TypeDescription struct {
	Category    string             `json:"category"`
	Description string             `json:"description,omitempty"`
	AliasType   string             `json:"alias-type,omitempty"`
	Fields      []FieldDescription `json:"fields,omitempty"`
	ElementType string             `json:"element-type,omitempty"`
	KeyType     string             `json:"-"`
	ValueType   string             `json:"-"`
}

func SetAPIDescription(description *APIDescription) {
	apiDescription = description
}

func GetAPIDescription() *APIDescription {
	return apiDescription
}

func CompareFieldDescription(a FieldDescription, b FieldDescription) int {
	return strings.Compare(a.Name, b.Name)
}

func GetDescriptionFromHandler(basePath string) (string, error) {
	absPath := getHandlerFilePath(basePath)
	if !util.IsFile(absPath) {
		return "", fmt.Errorf("Unable to find file path to API Handler. Endpoint: '%s'. Expected path: '%s'.", basePath, absPath)
	}

	handlePattern := regexp.MustCompile(`Handle`)
	description, err := util.GetDescriptionFromFunction(absPath, handlePattern)
	if err != nil {
		return "", fmt.Errorf("Error while getting description from function: '%v'. Endpoint: '%s'.", err, basePath)
	}

	return description, nil
}

// Routes must be validated before calling Describe().
func Describe(routes []Route) (*APIDescription, error) {
	endpointMap := make(map[string]EndpointDescription)
	typeMap := make(map[string]TypeDescription)

	var errs error = nil
	var err error
	for _, route := range routes {
		apiRoute, ok := route.(*APIRoute)
		if !ok {
			continue
		}

		// Check if we have already found the description for an endpoint.
		if apiRoute.Description == "" {
			apiRoute.Description, err = GetDescriptionFromHandler(apiRoute.GetBasePath())
			if err != nil {
				errs = errors.Join(errs, err)
			}
		}

		// RequestType and ResponseType must be structs, so Fields will hold the type's information.
		input, _, typeMap, _, err := DescribeType(apiRoute.RequestType, false, typeMap, nil)
		if err != nil {
			errs = errors.Join(errs, err)
		}

		output, _, typeMap, _, err := DescribeType(apiRoute.ResponseType, false, typeMap, nil)
		if err != nil {
			errs = errors.Join(errs, err)
		}

		endpointMap[apiRoute.GetBasePath()] = EndpointDescription{
			Description:  apiRoute.Description,
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
			Input:        input.Fields,
			Output:       output.Fields,
		}
	}

	apiDescription := APIDescription{
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

// Given a type and a map of known type descriptions, DescribeType() returns the type description and typeID.
//   - Basic types (PODs) return their String() as both an Alias and typeID.
//   - Arrays and slices store the typeID of their element in ElementType.
//   - Maps store the key type as a string and the value as a typeID in KeyType and ValueType respectively.
//   - Structs have a Fields map describing each field, including embedded ones.
//     Non-embedded struct fields that do not have a JSON tag are skipped.
func DescribeType(customType reflect.Type, addType bool, typeMap map[string]TypeDescription, typeConversions map[string]string) (TypeDescription, string, map[string]TypeDescription, map[string]string, error) {
	if customType == nil {
		return TypeDescription{}, "", map[string]TypeDescription{}, map[string]string{}, fmt.Errorf("Unable to describe nil type.")
	}

	if typeMap == nil {
		typeMap = make(map[string]TypeDescription)
	}

	if typeConversions == nil {
		typeConversions = make(map[string]string)
	}

	originalTypeID, err := getTypeID(customType, typeConversions)
	if err != nil {
		return TypeDescription{}, "", map[string]TypeDescription{}, map[string]string{}, err
	}

	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	typeID, err := getTypeID(customType, typeConversions)
	if err != nil {
		return TypeDescription{}, "", map[string]TypeDescription{}, map[string]string{}, err
	}

	typeDescription, ok := typeMap[typeID]
	if ok {
		return typeDescription, originalTypeID, typeMap, typeConversions, nil
	}

	switch customType.Kind() {
	case reflect.Slice, reflect.Array:
		_, elemTypeID, _, _, err := DescribeType(customType.Elem(), true, typeMap, typeConversions)
		if err != nil {
			return TypeDescription{}, "", map[string]TypeDescription{}, map[string]string{}, err
		}

		typeDescription.Category = ArrayType
		typeDescription.ElementType = elemTypeID
	case reflect.Map:
		_, elemTypeID, _, _, err := DescribeType(customType.Elem(), true, typeMap, typeConversions)
		if err != nil {
			return TypeDescription{}, "", map[string]TypeDescription{}, map[string]string{}, err
		}

		_, keyTypeID, _, _, err := DescribeType(customType.Key(), true, typeMap, typeConversions)
		if err != nil {
			return TypeDescription{}, "", map[string]TypeDescription{}, map[string]string{}, err
		}

		typeDescription.Category = MapType
		typeDescription.KeyType = keyTypeID
		typeDescription.ValueType = elemTypeID
	case reflect.Struct:
		typeDescription.Category = StructType
		typeDescription.Fields, err = describeStructFields(customType, typeMap, typeConversions)
		if err != nil {
			return TypeDescription{}, "", map[string]TypeDescription{}, map[string]string{}, err
		}

		if !addType {
			break
		}

		if customType.Name() == "" {
			break
		}

		if customType.PkgPath() == "" {
			break
		}

		descriptions, err := util.GetAllTypeDescriptionsFromPackage(customType.PkgPath())
		if err != nil {
			return TypeDescription{}, "", map[string]TypeDescription{}, map[string]string{}, err
		}

		description, ok := descriptions[customType.Name()]
		if ok {
			typeDescription.Description = description
		}
	default:
		// Handle built-in types.
		typeDescription.Category = AliasType

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

func describeStructFields(customType reflect.Type, typeMap map[string]TypeDescription, typeConversions map[string]string) ([]FieldDescription, error) {
	fieldDescriptions := make([]FieldDescription, 0)

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
			fieldDescription, fieldTypeID, _, _, err := DescribeType(field.Type, true, typeMap, typeConversions)
			if err != nil {
				return []FieldDescription{}, err
			}

			// If the embedded type is a struct, merge its fields into the current struct.
			if len(fieldDescription.Fields) > 0 {
				for _, description := range fieldDescription.Fields {
					if skipField(description.Name) {
						continue
					}

					fieldDescriptions = append(fieldDescriptions, description)
				}
			} else if fieldDescription.Category == AliasType {
				// Store basic embedded types under the current JSON tag.
				description := FieldDescription{
					Name: jsonTag,
					Type: fieldDescription.AliasType,
				}

				fieldDescriptions = append(fieldDescriptions, description)
			} else {
				// Store non-basic embedded types under the current JSON tag using the typeID.
				description := FieldDescription{
					Name: jsonTag,
					Type: fieldTypeID,
				}

				fieldDescriptions = append(fieldDescriptions, description)
			}

			continue
		}

		// Non-embedded fields must have a JSON field name.
		jsonTag = util.JSONFieldNameFull(field, false)
		if jsonTag == "" {
			continue
		}

		typeDescription, typeID, _, _, err := DescribeType(field.Type, true, typeMap, typeConversions)
		if err != nil {
			return []FieldDescription{}, err
		}

		if typeDescription.Category == AliasType {
			fieldDescription := FieldDescription{
				Name: jsonTag,
				Type: typeDescription.AliasType,
			}

			fieldDescriptions = append(fieldDescriptions, fieldDescription)
		} else {
			fieldDescription := FieldDescription{
				Name: jsonTag,
				Type: typeID,
			}

			fieldDescriptions = append(fieldDescriptions, fieldDescription)
		}
	}

	slices.SortFunc(fieldDescriptions, CompareFieldDescription)

	return fieldDescriptions, nil
}

func skipField(name string) bool {
	for _, pattern := range skipDescriptionPatterns {
		if pattern.MatchString(name) {
			return true
		}
	}

	return false
}

func getHandlerFilePath(basePath string) string {
	if strings.HasPrefix(basePath, "/") {
		basePath = strings.TrimPrefix(basePath, "/")
	}

	return util.ShouldAbs(filepath.Join(util.ShouldGetThisDir(), "..", basePath)) + ".go"
}
