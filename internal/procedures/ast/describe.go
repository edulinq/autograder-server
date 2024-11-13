package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

func GetAPIDescriptions() map[string]map[string]string {
	paths, err := util.FindDirents("", util.ShouldAbs(filepath.Join(common.ShouldGetCWD(), "internal", "api")), true, false, false)
	if err != nil {
		log.Fatal("Error finding internal/api dirents: '%v'.", err)
	}

	fset := token.NewFileSet()
	apiDescriptions := make(map[string]map[string]string)

	for _, path := range paths {
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			log.Warn("Unable to parse file '%s': %v", path, err)
			continue
		}

		for _, decl := range node.Decls {
			function, isFunction := decl.(*ast.FuncDecl)

			if isFunction && function.Name.IsExported() && function.Doc != nil {
				if len(function.Name.Name) > 6 && function.Name.Name[:6] == "Handle" {
					_, exists := apiDescriptions[function.Name.Name]
					if !exists {
						apiDescriptions[function.Name.Name] = make(map[string]string)
					}

					trimmedPath := path
					index := strings.Index(path, "internal/api")
					if index != -1 {
						trimmedPath = path[index:]
						trimmedPath = strings.ReplaceAll(trimmedPath, "internal/api/", "")
					}

					apiDescriptions[function.Name.Name][trimmedPath] = function.Doc.Text()
				}
			}
		}
	}

	return apiDescriptions
}
