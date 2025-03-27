package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/edulinq/autograder/internal/log"
)

const DEFAULT_PREFIX = ""
const DEFAULT_INDENT = "    "

// The target must be a pointer.
func JSONFromFile(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Could not open JSON file (%s): '%w'.", path, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Could not read JSON file (%s): '%w'.", path, err)
	}

	err = json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("Could not unmarshal JSON file (%s): '%w'.", path, err)
	}

	return nil
}

func MustJSONFromString(data string, target any) {
	err := JSONFromString(data, target)
	if err != nil {
		log.Fatal("Failed to convert JSON to object.", log.NewAttr("data", data), err)
	}
}

func JSONFromString(data string, target any) error {
	return JSONFromBytes([]byte(data), target)
}

func MustJSONFromBytes(data []byte, target any) {
	err := JSONFromBytes(data, target)
	if err != nil {
		log.Fatal("Failed to convert JSON to object.", log.NewAttr("data", data), err)
	}
}

func JSONFromBytes(data []byte, target any) error {
	err := json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("Could not unmarshal JSON bytes/string (%s): '%w'.", string(data), err)
	}

	return nil
}

func MustJSONMapFromString(data string) map[string]any {
	result, err := JSONMapFromString(data)
	if err != nil {
		log.Fatal("Could not create JSON map from string.", err, log.NewAttr("string", data))
	}

	return result
}

func JSONMapFromString(data string) (map[string]any, error) {
	target := make(map[string]any)

	err := json.Unmarshal([]byte(data), &target)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal JSON map from string (%s): '%w'.", data, err)
	}

	return target, nil
}

func MustToJSON(data any) string {
	text, err := ToJSON(data)
	if err != nil {
		log.Fatal("Failed to convert object to JSON.", log.NewAttr("data", data), err)
	}

	return text
}

func ToJSON(data any) (string, error) {
	return ToJSONIndentCustom(data, "", "")
}

func MustToJSONIndent(data any) string {
	text, err := ToJSONIndent(data)
	if err != nil {
		log.Fatal("Failed to convert object to JSON.", log.NewAttr("data", data), err)
	}

	return text
}

func ToJSONIndent(data any) (string, error) {
	return ToJSONIndentCustom(data, DEFAULT_PREFIX, DEFAULT_INDENT)
}

func ToJSONIndentCustom(data any, prefix string, indent string) (string, error) {
	return unmarshal(data, prefix, indent)
}

func ToJSONFile(data any, path string) error {
	return ToJSONFileIndentCustom(data, path, "", "")
}

func ToJSONFileIndent(data any, path string) error {
	return ToJSONFileIndentCustom(data, path, DEFAULT_PREFIX, DEFAULT_INDENT)
}

func ToJSONFileIndentCustom(data any, path string, prefix string, indent string) error {
	text, err := ToJSONIndentCustom(data, prefix, indent)
	if err != nil {
		return err
	}

	return WriteFile(text, path)
}

// Take a best shot at getting what the key would be for this the field in a JSON object.
func JSONFieldName(field reflect.StructField) string {
	name := field.Name

	tag := field.Tag.Get("json")
	if tag == "" {
		return name
	}

	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// This field is omitted from JSON.
		if part == "-" {
			return ""
		}

		// This special option is allowed.
		if part == "omitempty" {
			continue
		}

		return part
	}

	return name
}

// The only place we call json.Marshal*.
func unmarshal(data any, prefix string, indent string) (string, error) {
	var bytes []byte
	var err error

	if (prefix == "") && (indent == "") {
		bytes, err = json.Marshal(data)
	} else {
		bytes, err = json.MarshalIndent(data, prefix, indent)
	}

	if err != nil {
		// Explicitly use Go-Syntax (%#v) to avoid loops with overwritten String() methods.
		return "", fmt.Errorf("Could not marshal object ('%#v'): '%w'.", data, err)
	}

	return string(bytes), nil
}

// Marshal an enum using a map of possible values in a way that can be used for MarshalJSON().
func MarshalEnum[T comparable](value T, mapping map[T]string) ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(mapping[value])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// Unmarshal an enum using a map of possible values in a way that can be used for UnmarshalJSON().
func UnmarshalEnum[T comparable](data []byte, mapping map[string]T, lowerCase bool) (*T, error) {
	var temp string

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return nil, err
	}

	if lowerCase {
		temp = strings.ToLower(temp)
	}

	value, ok := mapping[temp]
	if !ok {
		return nil, fmt.Errorf("Found invalid value '%s' for enum %T.", temp, value)
	}

	return &value, nil
}

// Change the type of an object by serializing it to JSON and then deserializing as the new type.
func JSONTransformTypes[T any](rawValue any, defaultValue T) (T, error) {
	jsonString, err := ToJSON(rawValue)
	if err != nil {
		return defaultValue, err
	}

	err = JSONFromString(jsonString, &defaultValue)
	return defaultValue, err
}

func MustJSONTransformTypes[T any](rawValue any, defaultValue T) T {
	value, err := JSONTransformTypes(rawValue, defaultValue)
	if err != nil {
		log.Fatal("Failed to transform type via JSON serialization.", err)
	}

	return value
}

// Format a JSON object string.
func MustFormatJSONObject(text string) string {
	var object map[string]any
	MustJSONFromString(text, &object)
	return MustToJSON(object)
}

func MustFormatJSONObjectIndent(text string) string {
	var object map[string]any
	MustJSONFromString(text, &object)
	return MustToJSONIndent(object)
}

func MustToJSONMap(data any) map[string]any {
	dataJSONMap, err := ToJSONMap(data)
	if err != nil {
		log.Fatal("Failed to convert data to a JSON map.", err)
	}

	return dataJSONMap
}

func ToJSONMap(data any) (map[string]any, error) {
	dataJSON, err := ToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert data to JSON: '%v'.", err)
	}

	dataJSONMap, err := JSONMapFromString(dataJSON)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert JSON string to JSON map: '%v'.", err)
	}

	return dataJSONMap, nil
}
