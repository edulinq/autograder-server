package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

// API Description will be empty until RunServer() is called.
var apiDescription APIDescription

type APIDescription struct {
	Endpoints map[string]EndpointDescription `json:"endpoints"`
}

type EndpointDescription struct {
	RequestType  string `json:"request-type"`
	ResponseType string `json:"response-type"`
	Description  string `json:"description"`
}

func SetAPIDescription(description APIDescription) {
	apiDescription = description
}

func GetAPIDescription() APIDescription {
	return apiDescription
}

func GetDescriptionFromHandler(basePath string) (string, error) {
	absPath := makeAbsLocalAPIPath(basePath) + ".go"
	if !util.IsFile(absPath) {
		return "", fmt.Errorf("Unable to find file path to API Handler. Endpoint: '%s'. Expected path: '%s'.", basePath, absPath)
	}

	fset := token.NewFileSet()
	node, parseErr := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if parseErr != nil {
		return "", fmt.Errorf("Error while parsing file to get API description '%v'. Endpoint: '%s'.", parseErr, basePath)
	}

	for _, decl := range node.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if function.Name == nil {
			continue
		}

		if function.Doc == nil {
			continue
		}

		if strings.Contains(function.Name.Name, "Handle") {
			return strings.TrimSpace(function.Doc.Text()), nil
		}
	}

	return "", fmt.Errorf("Unable to find a description for an endpoint: '%s'.", basePath)
}

func makeAbsLocalAPIPath(suffix string) string {
	if strings.HasPrefix(suffix, "/") {
		suffix = strings.TrimPrefix(suffix, "/")
	}

	return util.ShouldAbs(filepath.Join(util.ShouldGetThisDir(), "..", suffix))
}
