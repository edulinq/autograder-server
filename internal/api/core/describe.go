package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/log"
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

type APIDescription struct {
	Endpoints map[string]EndpointDescription `json:"endpoints"`
	Types     map[string]TypeDescription     `json:"types"`
}

type EndpointDescription struct {
	Description  string                 `json:"description"`
	RequestType  string                 `json:"-"`
	ResponseType string                 `json:"-"`
	Input        []FieldDescription     `json:"input"`
	Output       []BaseFieldDescription `json:"output"`
}

type BaseFieldDescription struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type FieldDescription struct {
	BaseFieldDescription

	Required bool `json:"required,omitzero"`
}

type BaseTypeDescription struct {
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	AliasType   string `json:"alias-type,omitempty"`
	ElementType string `json:"element-type,omitempty"`
	KeyType     string `json:"-"`
	ValueType   string `json:"-"`
}

type TypeDescription struct {
	BaseTypeDescription

	Fields []BaseFieldDescription `json:"fields,omitempty"`
}

type FullTypeDescription struct {
	BaseTypeDescription

	Fields []FieldDescription `json:"fields,omitempty"`
}

type StructDescription map[string]string

type TypeInfoCache struct {
	TypeMap         map[string]FullTypeDescription
	TypeConversions map[string]string
	KnownPackages   map[string]StructDescription
}

func GetAPIDescription(forceCompute bool) (*APIDescription, error) {
	if !forceCompute {
		apiDescription, err := getCachedAPIDescription()
		if err != nil {
			return nil, fmt.Errorf("Failed to get cached API description: '%v'.", err)
		}

		if apiDescription != nil {
			return apiDescription, nil
		}
	}

	routes := GetAPIRoutes()
	if routes == nil || len(*routes) == 0 {
		return nil, fmt.Errorf("Unable to describe API endpoints when the cached routes are empty.")
	}

	apiDescription, err := DescribeRoutes(*routes)
	if err != nil {
		return nil, fmt.Errorf("Failed to describe API endpoints: '%v'.", err)
	}

	return apiDescription, nil
}

func getCachedAPIDescription() (*APIDescription, error) {
	apiDescriptionPath, err := util.GetAPIDescriptionFilepath()
	if err != nil {
		log.Warn("Unable to get cached API description.", err)
		return nil, nil
	}

	var apiDescription APIDescription
	err = util.JSONFromFile(apiDescriptionPath, &apiDescription)
	if err != nil {
		return nil, fmt.Errorf("Failed to get API description from file: '%v'.", err)
	}

	return &apiDescription, nil
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

func SimplifyTypeMap(fullTypeMap map[string]FullTypeDescription) map[string]TypeDescription {
	typeMap := make(map[string]TypeDescription, len(fullTypeMap))

	for typeName, description := range fullTypeMap {
		fields := make([]BaseFieldDescription, 0, len(description.Fields))

		for _, field := range description.Fields {
			fields = append(fields, field.BaseFieldDescription)
		}

		typeMap[typeName] = TypeDescription{
			BaseTypeDescription: description.BaseTypeDescription,
			Fields:              fields,
		}
	}

	return typeMap
}

// Routes must be validated before calling DescribeRoutes().
func DescribeRoutes(routes []Route) (*APIDescription, error) {
	endpointMap := make(map[string]EndpointDescription)
	info := TypeInfoCache{
		TypeMap: make(map[string]FullTypeDescription),
	}

	var errs error = nil
	var err error = nil

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
		input, _, info, err := DescribeType(apiRoute.RequestType, false, info)
		if err != nil {
			errs = errors.Join(errs, err)
		}

		output, _, info, err := DescribeType(apiRoute.ResponseType, false, info)
		if err != nil {
			errs = errors.Join(errs, err)
		}

		outputBaseFields := make([]BaseFieldDescription, 0, len(output.Fields))
		for _, field := range output.Fields {
			outputBaseFields = append(outputBaseFields, field.BaseFieldDescription)
		}

		endpointMap[apiRoute.GetBasePath()] = EndpointDescription{
			Description:  apiRoute.Description,
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
			Input:        input.Fields,
			Output:       outputBaseFields,
		}
	}

	finalTypeDescription := SimplifyTypeMap(info.TypeMap)

	apiDescription := APIDescription{
		Endpoints: endpointMap,
		Types:     finalTypeDescription,
	}

	return &apiDescription, errs
}

// Type conversions is an optional parameter that ensures that type IDs are not ambiguous.
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
func DescribeType(customType reflect.Type, addType bool, info TypeInfoCache) (FullTypeDescription, string, TypeInfoCache, error) {
	if customType == nil {
		return FullTypeDescription{}, "", TypeInfoCache{}, fmt.Errorf("Unable to describe nil type.")
	}

	if info.TypeMap == nil {
		info.TypeMap = make(map[string]FullTypeDescription)
	}

	if info.TypeConversions == nil {
		info.TypeConversions = make(map[string]string)
	}

	if info.KnownPackages == nil {
		info.KnownPackages = make(map[string]StructDescription)
	}

	originalTypeID, err := getTypeID(customType, info.TypeConversions)
	if err != nil {
		return FullTypeDescription{}, "", TypeInfoCache{}, err
	}

	if customType.Kind() == reflect.Pointer {
		customType = customType.Elem()
	}

	typeID, err := getTypeID(customType, info.TypeConversions)
	if err != nil {
		return FullTypeDescription{}, "", TypeInfoCache{}, err
	}

	typeDescription, ok := info.TypeMap[typeID]
	if ok {
		return typeDescription, originalTypeID, info, nil
	}

	switch customType.Kind() {
	case reflect.Array, reflect.Slice:
		_, elemTypeID, _, err := DescribeType(customType.Elem(), true, info)
		if err != nil {
			return FullTypeDescription{}, "", TypeInfoCache{}, err
		}

		typeDescription.Category = ArrayType
		typeDescription.ElementType = elemTypeID
	case reflect.Map:
		_, elemTypeID, _, err := DescribeType(customType.Elem(), true, info)
		if err != nil {
			return FullTypeDescription{}, "", TypeInfoCache{}, err
		}

		_, keyTypeID, _, err := DescribeType(customType.Key(), true, info)
		if err != nil {
			return FullTypeDescription{}, "", TypeInfoCache{}, err
		}

		typeDescription.Category = MapType
		typeDescription.KeyType = keyTypeID
		typeDescription.ValueType = elemTypeID
	case reflect.Struct:
		typeDescription.Category = StructType
		typeDescription.Fields, err = describeStructFields(customType, info)
		if err != nil {
			return FullTypeDescription{}, "", TypeInfoCache{}, err
		}

		if !addType {
			break
		}

		description, err := describeStruct(customType, info.KnownPackages)
		if err != nil {
			return FullTypeDescription{}, "", TypeInfoCache{}, err
		}

		typeDescription.Description = description
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
		info.TypeMap[typeID] = typeDescription
	}

	return typeDescription, originalTypeID, info, nil
}

func describeStructFields(customType reflect.Type, info TypeInfoCache) ([]FieldDescription, error) {
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

		requiredField, err := isFieldRequired(field)
		if err != nil {
			return nil, err
		}

		// Handle embedded fields.
		if field.Anonymous {
			fieldDescription, fieldTypeID, _, err := DescribeType(field.Type, false, info)
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
					BaseFieldDescription: BaseFieldDescription{
						Name: jsonTag,
						Type: fieldDescription.AliasType,
					},
					Required: requiredField,
				}

				fieldDescriptions = append(fieldDescriptions, description)
			} else {
				// Store non-basic embedded types under the current JSON tag using the typeID.
				description := FieldDescription{
					BaseFieldDescription: BaseFieldDescription{
						Name: jsonTag,
						Type: fieldTypeID,
					},
					Required: requiredField,
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

		typeDescription, typeID, _, err := DescribeType(field.Type, true, info)
		if err != nil {
			return []FieldDescription{}, err
		}

		if typeDescription.Category == AliasType {
			fieldDescription := FieldDescription{
				BaseFieldDescription: BaseFieldDescription{
					Name: jsonTag,
					Type: typeDescription.AliasType,
				},
				Required: requiredField,
			}

			fieldDescriptions = append(fieldDescriptions, fieldDescription)
		} else {
			fieldDescription := FieldDescription{
				BaseFieldDescription: BaseFieldDescription{
					Name: jsonTag,
					Type: typeID,
				},
				Required: requiredField,
			}

			fieldDescriptions = append(fieldDescriptions, fieldDescription)
		}
	}

	slices.SortFunc(fieldDescriptions, CompareFieldDescription)

	return fieldDescriptions, nil
}

func describeStruct(customType reflect.Type, knownPackages map[string]StructDescription) (string, error) {
	if customType.Name() == "" {
		return "", nil
	}

	if customType.PkgPath() == "" {
		return "", nil
	}

	var err error

	descriptions, ok := knownPackages[customType.PkgPath()]
	if !ok {
		descriptions, err = util.GetAllTypeDescriptionsFromPackage(customType.PkgPath())
		if err != nil {
			return "", err
		}

		knownPackages[customType.PkgPath()] = descriptions
	}

	description, ok := descriptions[customType.Name()]
	if !ok {
		return "", nil
	}

	return description, nil
}

func isFieldRequired(field reflect.StructField) (bool, error) {
	tag, ok := field.Tag.Lookup("required")
	if !ok {
		return false, nil
	}

	if tag == "" {
		return true, nil
	}

	return false, fmt.Errorf("Unexpected required tag value. Expected: '', Actual: '%s'.", tag)
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
