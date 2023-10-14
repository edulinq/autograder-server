package api

import (
    "fmt"
    "net/http"
    "reflect"
    "regexp"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

var routes = []route{
    newRedirect("GET", ``, `/static/index.html`),
    newRedirect("GET", `/`, `/static/index.html`),
    newRedirect("GET", `/index.html`, `/static/index.html`),

    newRoute("GET", `/static`, handleStatic),
    newRoute("GET", `/static/.*`, handleStatic),

    newAPIRoute(`/api/v02/user/get`, handleUserGet),
}

// Handlers that internally handle and log errors should return nil and ensure that responses are written.
type RouteHandler func(response http.ResponseWriter, request *http.Request) error;

// A handler specifically for API endpoints.
// The first return value will be encoded as the "content" field on the response.
// The handler should take in an APIRequest derived type.
// We will do some reflection around this type to ensure the request JSON is deserialized into it.
// Thus alias is not actually used (any and reflection are used), but shows what the structure is.
type APIHandler func(*any) (*any, *APIError);

// Objects that have been reflexively verifed.
// Once validated, callers should feel safe calling reflection methods on these.
type ValidAPIHandler any;
// This is a pointer to a request.
type ValidAPIRequest any;

// Inspired by https://benhoyt.com/writings/go-routing/
type route struct {
    method string
    regex *regexp.Regexp
    handler RouteHandler
}

func StartServer() {
    var port = config.WEB_PORT.GetInt();

    log.Info().Msgf("Serving on %d.", port);

    err := http.ListenAndServe(fmt.Sprintf(":%d", port), http.HandlerFunc(serve));
    if (err != nil) {
        log.Fatal().Err(err).Msg("Server stopped.");
    }
}

func newRoute(method string, pattern string, handler RouteHandler) route {
    return route{method, regexp.MustCompile("^" + pattern + "$"), handler};
}

func newRedirect(method string, pattern string, target string) route {
    redirectFunc := func(response http.ResponseWriter, request *http.Request) error {
        return handleRedirect(target, response, request);
    };
    return route{method, regexp.MustCompile("^" + pattern + "$"), redirectFunc};
}

func newAPIRoute[T APIRequest](pattern string, apiHandler any) route {
    handler := func(response http.ResponseWriter, request *http.Request) (err error) {
        // Recover from any panic.
        defer func() {
            value := recover();
            if (value == nil) {
                return;
            }

            log.Error().Any("value", value).Str("endpoint", request.URL.Path).
                    Msg("Recovered from a panic when handling an API endpoint.");
            apiErr := NewBareInternalError("-501", request.URL.Path, "Recovered from a panic when handling an API endpoint.").
                    Add("value", value);

            err = sendAPIResponse(nil, response, nil, apiErr, false);
        }();

        err = handleAPIEndpoint(response, request, apiHandler);

        return err;
    }

    return route{"POST", regexp.MustCompile("^" + pattern + "$"), handler};
}

func handleRedirect(target string, response http.ResponseWriter, request *http.Request) error {
    http.Redirect(response, request, target, 301);
    return nil;
}

func handleAPIEndpoint(response http.ResponseWriter, request *http.Request, apiHandler any) error {
    // Ensure the handler looks good.
    validAPIHandler, apiErr := validateAPIHandler(request.URL.Path, apiHandler);
    if (apiErr != nil) {
        return apiErr;
    }

    // Get the actual request.
    apiRequest, apiErr := createAPIRequest(request, validAPIHandler);
    if (apiErr != nil) {
        return sendAPIResponse(nil, response, nil, apiErr, false);
    }

    // Execute the handler.
    apiResponse, apiErr := callHandler(apiHandler, apiRequest);

    return sendAPIResponse(apiRequest, response, apiResponse, apiErr, false);
}

// Send out the result from an API call.
// If the APIError is not null, then it will be sent and no content will be sent.
// Otherwise, send the content in the response's "content" field.
// |hardFail| controls whether we should try to wrap an error and call this method again (so we don't infinite loop),
// most callers should set it to false.
func sendAPIResponse(apiRequest ValidAPIRequest, response http.ResponseWriter, content any, apiErr *APIError, hardFail bool) error {
    var apiResponse *APIResponse = nil;

    if (apiErr != nil) {
        apiResponse = apiErr.ToResponse();

        // This is the last interaction we will have with this error, log it.
        apiErr.Log();
    } else {
        apiResponse = NewAPIResponse(apiRequest, content);
    }

    payload, err := util.ToJSON(apiResponse);
    if (err != nil) {
        apiErr = NewBareInternalError("-531", "", "Could not serialize API response.").Err(err);
        apiResponse = apiErr.ToResponse();

        if (hardFail) {
            log.Error().Err(err).Any("request", apiRequest).Any("response", apiResponse).Any("api-error", apiErr).
                    Msg("Failed to encode API result as JSON, hard failing.");

            payload, _ = util.ToJSON(apiResponse);
        } else {
            return sendAPIResponse(apiRequest, response, nil, apiErr, true);
        }
    }

    response.WriteHeader(apiResponse.HTTPStatus);

    _, err = fmt.Fprint(response, payload);
    if (err != nil) {
        http.Error(response, "Server Failed to Send Response - Contact Admins", http.StatusInternalServerError);
        log.Error().Err(err).Str("payload", payload).Msg("Failed to write final payload to http writer.");
        return fmt.Errorf("Could not write API response payload: '%w'.", err);
    }

    return nil;
}

// Reflexively create an API request for the handler from the content of the POST request.
func createAPIRequest(request *http.Request, apiHandler ValidAPIHandler) (ValidAPIRequest, *APIError) {
    endpoint := request.URL.Path;

    // Allocate memory for the request.
    apiRequest, apiErr := allocateAPIRequest(endpoint, apiHandler);
    if (apiErr != nil) {
        return nil, apiErr;
    }

    // TODO(eriq): Handle Files

    // Get the text from the POST.
    textContent := request.PostFormValue(API_REQUEST_CONTENT_KEY);
    if (textContent == "") {
        return nil, NewBareBadRequestError(endpoint,
                fmt.Sprintf("JSON payload for POST form key '%s' is empty.", API_REQUEST_CONTENT_KEY));
    }

    // Unmarshal the JSON.
    err := util.JSONFromString(textContent, apiRequest);
    if (err != nil) {
        return nil, NewBareBadRequestError(endpoint,
                fmt.Sprintf("JSON payload for POST form key '%s' is not valid JSON.", API_REQUEST_CONTENT_KEY)).
                Err(err);
    }

    // Validate the request.
    apiErr = ValidateAPIRequest(apiRequest, endpoint);
    if (apiErr != nil) {
        return nil, apiErr;
    }

    return ValidAPIRequest(apiRequest), nil;
}

// Reflexively call the API handler with the request.
func callHandler(apiHandler ValidAPIHandler, apiRequest ValidAPIRequest) (any, *APIError) {
    input := []reflect.Value{reflect.ValueOf(apiRequest)};
    output := reflect.ValueOf(apiHandler).Call(input);

    response := output[0].Interface();
    apiErr := output[1].Interface().(*APIError);

    return response, apiErr;
}

// Reflexively allocate a new object to hold an API request that can be used on the given handler.
func allocateAPIRequest(endpoint string, apiHandler ValidAPIHandler) (any, *APIError) {
    reflectType := reflect.TypeOf(apiHandler);
    argumentType := reflectType.In(0).Elem();
    requestPointer := reflect.New(argumentType).Interface();

    return requestPointer, nil;
}

// Reflexively ensure that the api handler is of the correct type/format (e.g. looks like APIHandler).
// Once you have a ValidAPIHandler, there is no need to check before doing reflection operations.
func validateAPIHandler(endpoint string, apiHandler any) (ValidAPIHandler, *APIError) {
    reflectValue := reflect.ValueOf(apiHandler);
    reflectType := reflect.TypeOf(apiHandler);

    if (reflectValue.Kind() != reflect.Func) {
        return nil, NewBareInternalError("-521", endpoint, "API handler is not a function.").
                Add("kind", reflectValue.Kind().String());
    }

    if (reflectType.NumIn() != 1) {
        return nil, NewBareInternalError("-522", endpoint, "API handler does not have exactly 1 argument.").
                Add("num-in", reflectType.NumIn());
    }
    argumentType := reflectType.In(0);

    if (argumentType.Kind() != reflect.Pointer) {
        return nil, NewBareInternalError("-523", endpoint, "API handler's argument is not a pointer.").
                Add("kind", argumentType.Kind().String());
    }

    if (reflectType.NumOut() != 2) {
        return nil, NewBareInternalError("-524", endpoint, "API handler does not return exactly 2 arguments.").
                Add("num-out", reflectType.NumOut());
    }

    if (reflectType.Out(0).Kind() != reflect.Pointer) {
        return nil, NewBareInternalError("-525", endpoint, "API handler's first return value is not a pointer.").
                Add("kind", reflectType.Out(0).Kind().String());
    }

    if (reflectType.Out(1) != reflect.TypeOf((*APIError)(nil))) {
        return nil, NewBareInternalError("-526", endpoint, "API handler's second return value is a *APIError.").
                Add("type", reflectType.Out(1).String());
    }

    return ValidAPIHandler(apiHandler), nil;
}

func serve(response http.ResponseWriter, request *http.Request) {
    log.Debug().
        Str("method", request.Method).
        Str("url", request.URL.Path).
        Msg("");

    var route route;
    var match bool;

    for _, route = range routes {
        if (route.method != request.Method) {
            continue;
        }

        match = route.regex.MatchString(request.URL.Path);
        if (!match) {
            continue;
        }

        err := route.handler(response, request);
        if (err != nil) {
            log.Error().Err(err).Str("path", request.URL.Path).Msg("Handler had an error.");
            http.Error(response, "Server Error", http.StatusInternalServerError);
        }

        return;
    }

    http.NotFound(response, request);
}
