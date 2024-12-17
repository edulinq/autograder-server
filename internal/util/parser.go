package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

// This function takes a file path and a pattern that mathes the name of the target function.
// Returns the comment attached to the target function.
// Errors occur when the file or target function cannot be found or if the target function does not have a comment attached.
func GetDescriptionFromFunction(path string, functionNamePattern *regexp.Regexp) (string, error) {
	if !IsFile(path) {
		return "", fmt.Errorf("Unable to find file path to function. Expected path: '%s'.", path)
	}

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("Error while parsing file to get function description: '%v'.", err)
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

		if functionNamePattern.MatchString(function.Name.Name) {
			return strings.TrimSpace(function.Doc.Text()), nil
		}
	}

	return "", fmt.Errorf("Unable to find a description using the pattern '%s' for a function located in '%s'.", functionNamePattern.String(), path)
}
