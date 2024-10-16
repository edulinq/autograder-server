package core

import (
	"net/http"
	"reflect"
	"regexp"

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
	Description string
	Request     reflect.Type
	Response    reflect.Type
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

func (this *APIRoute) GetDescription() string {
	return this.Description
}

func NewBaseRoute(method string, basePath string, handler RouteHandler) *BaseRoute {
	return &BaseRoute{
		Method:   method,
		BasePath: basePath,
		Regex:    regexp.MustCompile("^" + NewEndpoint(basePath) + "$"),
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
		Regex:    regexp.MustCompile("^" + NewEndpoint(basePath) + "$"),
		Handler:  redirectFunc,
	}
}

func NewAPIRoute(basePath string, apiHandler any, description string) *APIRoute {
	var endpointPath string

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

		endpointPath = request.URL.Path

		err = handleAPIEndpoint(response, request, apiHandler)

		return err
	}

	_, requestType, responseType, err := validateAPIHandler(endpointPath, apiHandler)
	if err != nil {
		log.Warn("Error while validating API handler.", err, log.NewAttr("endpoint", endpointPath))
	}

	return &APIRoute{
		BaseRoute: BaseRoute{
			Method:   "POST",
			BasePath: basePath,
			Regex:    regexp.MustCompile("^" + NewEndpoint(basePath) + "$"),
			Handler:  handler,
		},
		Description: description,
		Request:     requestType,
		Response:    responseType,
	}
}
