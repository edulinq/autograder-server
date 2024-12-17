package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

func GetDescriptionFromFunction(absPath string, re *regexp.Regexp) (string, error) {
	if !IsFile(absPath) {
		return "", fmt.Errorf("Unable to find file path to function. Expected path: '%s'.", absPath)
	}

	fset := token.NewFileSet()
	node, parseErr := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if parseErr != nil {
		return "", fmt.Errorf("Error while parsing file to get function description: '%v'.", parseErr)
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

		if re.MatchString(function.Name.Name) {
			return strings.TrimSpace(function.Doc.Text()), nil
		}
	}

	return "", fmt.Errorf("Unable to find a description using the pattern '%s' for a function located in '%s'.", re.String(), absPath)
}
