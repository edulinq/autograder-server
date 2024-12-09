package core

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/edulinq/autograder/internal/exit"
	"github.com/edulinq/autograder/internal/log"
)

type Route interface {
	GetMethod() string
	GetRegex() *regexp.Regexp
	GetBasePath() string
	Handle(response http.ResponseWriter, request *http.Request) error
}

// Inspired by https://benhoyt.com/writings/go-routing/
type BaseRoute struct {
	Method   string
	BasePath string
	Regex    *regexp.Regexp
	Handler  RouteHandler
}

type APIRoute struct {
	BaseRoute
	RequestType  reflect.Type
	ResponseType reflect.Type
	Description  string
}

func (this *BaseRoute) GetMethod() string {
	return this.Method
}

func (this *BaseRoute) GetRegex() *regexp.Regexp {
	return this.Regex
}

func (this *BaseRoute) GetBasePath() string {
	return this.BasePath
}

func (this *BaseRoute) Handle(response http.ResponseWriter, request *http.Request) error {
	return this.Handler(response, request)
}

func NewBaseRoute(method string, basePath string, handler RouteHandler) *BaseRoute {
	return &BaseRoute{
		Method:   method,
		BasePath: basePath,
		Regex:    regexp.MustCompile("^" + MakeFullAPIPath(basePath) + "$"),
		Handler:  handler,
	}
}

func NewRedirect(method string, basePath string, target string) *BaseRoute {
	redirectFunc := func(response http.ResponseWriter, request *http.Request) error {
		return handleRedirect(target, response, request)
	}

	return &BaseRoute{
		Method:   method,
		BasePath: basePath,
		Regex:    regexp.MustCompile("^" + MakeFullAPIPath(basePath) + "$"),
		Handler:  redirectFunc,
	}
}

func MustNewAPIRoute(basePath string, apiHandler any) *APIRoute {
	handler := func(response http.ResponseWriter, request *http.Request) (err error) {
		// Recover from any panic.
		defer func() {
			value := recover()
			if value == nil {
				return
			}

			log.Error("Recovered from a panic when handling an API endpoint.",
				log.NewAttr("value", value), log.NewAttr("endpoint", request.URL.Path))
			apiErr := NewBareInternalError("-001", request.URL.Path, "Recovered from a panic when handling an API endpoint.").
				Add("value", value)

			err = sendAPIResponse(nil, response, nil, apiErr, false)
		}()

		err = handleAPIEndpoint(response, request, apiHandler)

		return err
	}

	fullPath := MakeFullAPIPath(basePath)

	_, requestType, responseType, err := validateAPIHandler(fullPath, apiHandler)
	if err != nil {
		log.FatalWithCode(exit.EXIT_SOFTWARE, "Error while validating API handler.", err, log.NewAttr("endpoint", fullPath))
	}

	absPath := makeAbsLocalAPIPath(basePath) + ".go"
	fset := token.NewFileSet()
	node, parseErr := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if parseErr != nil {
		TODODescriptions := "TODO: basePath: '" + basePath + "'. Could not find: " + absPath
		return &APIRoute{
			BaseRoute: BaseRoute{
				Method:   "POST",
				BasePath: basePath,
				Regex:    regexp.MustCompile("^" + fullPath + "$"),
				Handler:  handler,
			},
			RequestType:  requestType,
			ResponseType: responseType,
			Description:  TODODescriptions,
		}

		// log.FatalWithCode(exit.EXIT_SOFTWARE, "Error while parsing file to get API description.", parseErr, log.NewAttr("endpoint", absPath))
	}

	apiDescription := ""
	for _, decl := range node.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if strings.Contains(function.Name.Name, "Handle") {
			if function.Doc != nil {
				apiDescription = strings.TrimSpace(function.Doc.Text())
				break
			}
		}
	}

	return &APIRoute{
		BaseRoute: BaseRoute{
			Method:   "POST",
			BasePath: basePath,
			Regex:    regexp.MustCompile("^" + fullPath + "$"),
			Handler:  handler,
		},
		RequestType:  requestType,
		ResponseType: responseType,
		Description:  apiDescription,
	}
}
