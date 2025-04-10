package core

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

const (
	AliasType  = "alias"
	ArrayType  = "array"
	MapType    = "map"
	StructType = "struct"
)

// API Description will be empty until RunServer() is called.
var apiDescription APIDescription

type APIDescription struct {
	Endpoints map[string]EndpointDescription `json:"endpoints"`
	Types     map[string]FullTypeDescription `json:"types"`
}

type EndpointDescription struct {
	Description  string            `json:"description"`
	RequestType  string            `json:"-"`
	ResponseType string            `json:"-"`
	Input        []TypeDescription `json:"input"`
	Output       []TypeDescription `json:"output"`
}

type TypeDescription struct {
	FieldName string `json:"field-name"`
	FieldType string `json:"field-type"`
	// FieldDescription string `json:"field-description"`
}

type FullTypeDescription struct {
	Category    string            `json:"category"`
	Description string            `json:"description,omitempty"`
	AliasType   string            `json:"alias-type,omitempty"`
	Fields      []TypeDescription `json:"fields,omitempty"`
	ElementType string            `json:"element-type,omitempty"`
	KeyType     string            `json:"-"`
	ValueType   string            `json:"-"`
}

func SetAPIDescription(description APIDescription) {
	apiDescription = description
}

func GetAPIDescription() APIDescription {
	return apiDescription
}

func CompareTypeDescriptionPointer(a *TypeDescription, b *TypeDescription) int {
	if a == b {
		return 0
	}

	if a == nil {
		return 1
	}

	if b == nil {
		return -1
	}

	return CompareTypeDescription(*a, *b)
}

func CompareTypeDescription(a TypeDescription, b TypeDescription) int {
	return strings.Compare(a.FieldName, b.FieldName)
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

func getHandlerFilePath(basePath string) string {
	if strings.HasPrefix(basePath, "/") {
		basePath = strings.TrimPrefix(basePath, "/")
	}

	return util.ShouldAbs(filepath.Join(util.ShouldGetThisDir(), "..", basePath)) + ".go"
}
