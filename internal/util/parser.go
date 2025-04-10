package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

// This function takes a file path and a pattern that matches the name of the target function.
// Returns the comment attached to the first occurrence of the target function.
// Errors occur when the target function cannot be found.
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

		if functionNamePattern.MatchString(function.Name.Name) {
			if function.Doc == nil {
				return "", nil
			}

			return strings.TrimSpace(function.Doc.Text()), nil
		}
	}

	return "", fmt.Errorf("Unable to find a description using the pattern '%s' for a function located in '%s'.", functionNamePattern.String(), path)
}

func GetDescriptionFromType(packagePath string, typeName string) (string, error) {
	dirPath := getDirPathFromPackagePath(packagePath)

	filePaths, err := FindFiles("", dirPath)
	if err != nil {
		return "", fmt.Errorf("Unable to find file paths for the package path '%s': '%v'.", packagePath, err)
	}

	description, err := getDescriptionFromType(filePaths, typeName)
	if err != nil {
		return "", fmt.Errorf("Unable to get description for the type '%s' in '%s': '%v'.", typeName, packagePath, err)
	}

	return description, nil
}

func getDescriptionFromType(filePaths []string, typeName string) (string, error) {
	for _, path := range filePaths {
		if !IsFile(path) {
			continue
		}

		if !strings.HasSuffix(path, ".go") {
			continue
		}

		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			return "", fmt.Errorf("Error while parsing file to get function description: '%v'.", err)
		}

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			if genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if typeSpec.Name == nil {
					continue
				}

				if typeSpec.Name.Name == typeName {
					if genDecl.Doc == nil {
						return "", nil
					}

					return strings.TrimSpace(genDecl.Doc.Text()), nil
				}
			}
		}
	}

	return "", fmt.Errorf("Unable to find a description using the name '%s' for a type located in '%v'.", typeName, filePaths)
}
