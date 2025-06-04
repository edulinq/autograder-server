package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
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

// This function takes a non-empty package path and returns a map of custom type names to their description.
// The types of package paths accepted can be seen in GetDirPathFromCustomPackagePath().
func GetAllTypeDescriptionsFromPackage(packagePath string) (map[string]string, error) {
	dirPath := GetDirPathFromCustomPackagePath(packagePath)

	filePaths, err := FindFiles("", dirPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to find file paths for the package path '%s': '%v'.", packagePath, err)
	}

	descriptions, err := getDescriptionFromType(filePaths)
	if err != nil {
		return nil, fmt.Errorf("Unable to get descriptions for the package path '%s': '%v'.", packagePath, err)
	}

	return descriptions, nil
}

func GetFieldDescriptionsFromType(customType reflect.Type) (map[string]string, error) {
	dirPath := GetDirPathFromCustomPackagePath(customType.PkgPath())

	filePaths, err := FindFiles("", dirPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to find file paths for the package path '%s': '%v'.", customType.PkgPath(), err)
	}

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
			return nil, fmt.Errorf("Error while parsing file to get function description: '%v'.", err)
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

				if typeSpec.Name.Name != customType.Name() {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					return nil, fmt.Errorf("Type '%s' must be a struct type.", customType.Name())
				}

				descriptions := getFieldDescriptionsFromStructType(structType)

				return descriptions, nil
			}
		}
	}

	return nil, fmt.Errorf("Unable to find the type declaration for '%s' in the package path: '%s'.", customType.Name(), customType.PkgPath())
}

func getDescriptionFromType(filePaths []string) (map[string]string, error) {
	descriptions := make(map[string]string, 0)

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
			return nil, fmt.Errorf("Error while parsing file to get function description: '%v'.", err)
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

				if genDecl.Doc == nil {
					descriptions[typeSpec.Name.Name] = ""
					continue
				}

				descriptions[typeSpec.Name.Name] = strings.TrimSpace(genDecl.Doc.Text())
			}
		}
	}

	return descriptions, nil
}

func getFieldDescriptionsFromStructType(structType *ast.StructType) map[string]string {
	descriptions := make(map[string]string, 0)

	if structType == nil {
		return descriptions
	}

	if structType.Fields == nil {
		return descriptions
	}

	if structType.Fields.List == nil {
		return descriptions
	}

	for _, field := range structType.Fields.List {
		if field == nil {
			continue
		}

		if field.Names == nil {
			continue
		}

		fieldName := ""
		for _, name := range field.Names {
			if name == nil {
				continue
			}

			fieldName = name.Name
			break
		}

		if fieldName == "" {
			continue
		}

		if field.Doc == nil {
			descriptions[fieldName] = ""
			continue
		}

		descriptions[fieldName] = strings.TrimSpace(field.Doc.Text())
	}

	return descriptions
}
