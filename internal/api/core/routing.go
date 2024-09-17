package core

// The infrastructure for routing requests,
// mostly API requests.

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/edulinq/autograder/internal/api/static"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

// Handlers that internally handle and log errors should return nil and ensure that responses are written.
type RouteHandler func(response http.ResponseWriter, request *http.Request) error

// A handler specifically for API endpoints.
// The first return value will be encoded as the "content" field on the response.
// The handler should take in an APIRequest derived type.
// We will do some reflection around this type to ensure the request JSON is deserialized into it.
// Thus alias is not actually used (any and reflection are used), but shows what the structure is.
type APIHandler func(*any) (*any, *APIError)

// A handler that has been reflexively verified.
// Once validated, callers should feel safe calling reflection methods on this without extra checks.
type ValidAPIHandler any

// Post form key for request content.
const API_REQUEST_CONTENT_KEY = "content"

// Inspired by https://benhoyt.com/writings/go-routing/
type Route struct {
	method  string
	regex   *regexp.Regexp
	handler RouteHandler
}

const MAX_FORM_MEM_SIZE_BYTES = 20 * 1024 * 1024 // 20 MB

// Get a function to pass to http.HandlerFunc().
func GetRouteServer(routes *[]*Route) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ServeRoutes(routes, response, request)
	}
}

func ServeRoutes(routes *[]*Route, response http.ResponseWriter, request *http.Request) {
	log.Debug("Incoming Request", log.NewAttr("method", request.Method), log.NewAttr("url", request.URL.Path))

	if routes == nil {
		http.NotFound(response, request)
	}

	var i int
	var route *Route
	var match bool

	for i, route = range *routes {
		if route == nil {
			log.Warn("Found nil route.", log.NewAttr("index", i))
		}

		if route.method != request.Method {
			continue
		}

		match = route.regex.MatchString(request.URL.Path)
		if !match {
			continue
		}

		err := route.handler(response, request)
		if err != nil {
			log.Error("Handler had an error.", err, log.NewAttr("path", request.URL.Path))
			http.Error(response, "Server Error", http.StatusInternalServerError)
		}

		return
	}

	// If this path does not look like an API request and static fallback is enabled,
	// the try to match the path with a static path.
	if config.WEB_STATIC_FALLBACK.Get() && !strings.HasPrefix(request.URL.Path, CURRENT_PREFIX) {
		log.Trace("Attempting Static Fallback", log.NewAttr("method", request.Method), log.NewAttr("url", request.URL.Path))
		static.Handle(response, request)
		return
	}

	http.NotFound(response, request)
}

func NewRoute(method string, pattern string, handler RouteHandler) *Route {
	return &Route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}

func NewRedirect(method string, pattern string, target string) *Route {
	redirectFunc := func(response http.ResponseWriter, request *http.Request) error {
		return handleRedirect(target, response, request)
	}

	return &Route{method, regexp.MustCompile("^" + pattern + "$"), redirectFunc}
}

func NewAPIRoute(pattern string, apiHandler any) *Route {
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

	return &Route{"POST", regexp.MustCompile("^" + pattern + "$"), handler}
}

func handleRedirect(target string, response http.ResponseWriter, request *http.Request) error {
	http.Redirect(response, request, target, 301)
	return nil
}

func handleAPIEndpoint(response http.ResponseWriter, request *http.Request, apiHandler any) error {
	// Ensure the handler looks good.
	validAPIHandler, apiErr := validateAPIHandler(request.URL.Path, apiHandler)
	if apiErr != nil {
		return sendAPIResponse(nil, response, nil, apiErr, false)
	}

	// Get the actual request.
	apiRequest, apiErr := createAPIRequest(request, validAPIHandler)
	if apiErr != nil {
		return sendAPIResponse(nil, response, nil, apiErr, false)
	}
	defer CleanupAPIrequest(apiRequest)

	// Execute the handler.
	apiResponse, apiErr := callHandler(apiHandler, apiRequest)

	return sendAPIResponse(apiRequest, response, apiResponse, apiErr, false)
}

// Send out the result from an API call.
// If the APIError is not null, then it will be sent and no content will be sent.
// Otherwise, send the content in the response's "content" field.
// |hardFail| controls whether we should try to wrap an error and call this method again (so we don't infinite loop),
// most callers should set it to false.
func sendAPIResponse(apiRequest ValidAPIRequest, response http.ResponseWriter,
	content any, apiErr *APIError, hardFail bool) error {
	var apiResponse *APIResponse = nil

	if apiErr != nil {
		apiResponse = apiErr.ToResponse()

		// This is the last interaction we will have with this error, log it.
		apiErr.Log()
	} else {
		apiResponse = NewAPIResponse(apiRequest, content)
	}

	payload, err := util.ToJSON(apiResponse)
	if err != nil {
		apiErr = NewBareInternalError("-002", "", "Could not serialize API response.").Err(err)
		apiResponse = apiErr.ToResponse()

		if hardFail {
			log.Error("Failed to encode API result as JSON, hard failing.",
				err, log.NewAttr("request", apiRequest), log.NewAttr("response", apiResponse), log.NewAttr("api-error", apiErr))

			payload, _ = util.ToJSON(apiResponse)
		} else {
			return sendAPIResponse(apiRequest, response, nil, apiErr, true)
		}
	}

	// When in testing mode, allow cross-origin requests.
	if config.TESTING_MODE.Get() {
		response.Header().Set("Access-Control-Allow-Origin", "*")
	}

	response.WriteHeader(apiResponse.HTTPStatus)

	_, err = fmt.Fprint(response, payload)
	if err != nil {
		http.Error(response, "Server Failed to Send Response - Contact Admins", http.StatusInternalServerError)
		log.Error("Failed to write final payload to http writer.", err, log.NewAttr("payload", payload))
		return fmt.Errorf("Could not write API response payload: '%w'.", err)
	}

	return nil
}

// Reflexively create an API request for the handler from the content of the POST request.
func createAPIRequest(request *http.Request, apiHandler ValidAPIHandler) (ValidAPIRequest, *APIError) {
	endpoint := request.URL.Path

	// Allocate memory for the request.
	apiRequest, apiErr := allocateAPIRequest(endpoint, apiHandler)
	if apiErr != nil {
		return nil, apiErr
	}

	// If this request is multipart, then parse the form.
	if strings.Contains(strings.Join(request.Header["Content-Type"], " "), "multipart/form-data") {
		err := request.ParseMultipartForm(MAX_FORM_MEM_SIZE_BYTES)
		if err != nil {
			return nil, NewBareBadRequestError("-003", endpoint,
				fmt.Sprintf("POST request is improperly formatted.")).
				Err(err)
		}
	}

	// Get the text from the POST.
	textContent := request.PostFormValue(API_REQUEST_CONTENT_KEY)
	if textContent == "" {
		return nil, NewBareBadRequestError("-004", endpoint,
			fmt.Sprintf("JSON payload for POST form key '%s' is empty.", API_REQUEST_CONTENT_KEY))
	}

	// Unmarshal the JSON.
	err := util.JSONFromString(textContent, apiRequest)
	if err != nil {
		return nil, NewBareBadRequestError("-005", endpoint,
			fmt.Sprintf("JSON payload for POST form key '%s' is not valid JSON.", API_REQUEST_CONTENT_KEY)).
			Err(err)
	}

	// Validate the request.
	apiErr = ValidateAPIRequest(request, apiRequest, endpoint)
	if apiErr != nil {
		return nil, apiErr
	}

	return ValidAPIRequest(apiRequest), nil
}

// Reflexively call the API handler with the request.
func callHandler(apiHandler ValidAPIHandler, apiRequest ValidAPIRequest) (any, *APIError) {
	input := []reflect.Value{reflect.ValueOf(apiRequest)}
	output := reflect.ValueOf(apiHandler).Call(input)

	response := output[0].Interface()
	apiErr := output[1].Interface().(*APIError)

	return response, apiErr
}

// Reflexively allocate a new object to hold an API request that can be used on the given handler.
func allocateAPIRequest(endpoint string, apiHandler ValidAPIHandler) (any, *APIError) {
	reflectType := reflect.TypeOf(apiHandler)
	argumentType := reflectType.In(0).Elem()
	requestPointer := reflect.New(argumentType).Interface()

	return requestPointer, nil
}

// Reflexively ensure that the api handler is of the correct type/format (e.g. looks like APIHandler).
// Once you have a ValidAPIHandler, there is no need to check before doing reflection operations.
func validateAPIHandler(endpoint string, apiHandler any) (ValidAPIHandler, *APIError) {
	reflectValue := reflect.ValueOf(apiHandler)
	reflectType := reflect.TypeOf(apiHandler)

	if reflectValue.Kind() != reflect.Func {
		return nil, NewBareInternalError("-006", endpoint, "API handler is not a function.").
			Add("kind", reflectValue.Kind().String())
	}

	funcInfo := getFuncInfo(apiHandler)

	if reflectType.NumIn() != 1 {
		return nil, NewBareInternalError("-007", endpoint, "API handler does not have exactly 1 argument.").
			Add("num-in", reflectType.NumIn()).
			Add("function-info", funcInfo)
	}
	argumentType := reflectType.In(0)

	if argumentType.Kind() != reflect.Pointer {
		return nil, NewBareInternalError("-008", endpoint, "API handler's argument is not a pointer.").
			Add("kind", argumentType.Kind().String()).
			Add("function-info", funcInfo)
	}

	if reflectType.NumOut() != 2 {
		return nil, NewBareInternalError("-009", endpoint, "API handler does not return exactly 2 arguments.").
			Add("num-out", reflectType.NumOut()).
			Add("function-info", funcInfo)
	}

	if reflectType.Out(0).Kind() != reflect.Pointer {
		return nil, NewBareInternalError("-010", endpoint, "API handler's first return value is not a pointer.").
			Add("kind", reflectType.Out(0).Kind().String()).
			Add("function-info", funcInfo)
	}

	if reflectType.Out(1) != reflect.TypeOf((*APIError)(nil)) {
		return nil, NewBareInternalError("-011", endpoint, "API handler's second return value is a *APIError.").
			Add("type", reflectType.Out(1).String()).
			Add("function-info", funcInfo)
	}

	return ValidAPIHandler(apiHandler), nil
}

type funcRuntimeInfo struct {
	Name string
	File string
	Line int
}

func getFuncInfo(funcHandle any) *funcRuntimeInfo {
	info := funcRuntimeInfo{}

	reflectValue := reflect.ValueOf(funcHandle)
	if reflectValue.Kind() != reflect.Func {
		return &info
	}

	runtimeFunc := runtime.FuncForPC(reflectValue.Pointer())
	if runtimeFunc != nil {
		info.Name = runtimeFunc.Name()
		info.File, info.Line = runtimeFunc.FileLine(reflectValue.Pointer())
	}

	return &info
}
