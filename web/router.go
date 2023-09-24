package web

import (
    "fmt"
    "net/http"
    "regexp"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
)

var routes = []route{
    newRedirect("GET", ``, `/static/index.html`),
    newRedirect("GET", `/`, `/static/index.html`),
    newRedirect("GET", `/index.html`, `/static/index.html`),

    newRoute("GET", `/static`, handleStatic),
    newRoute("GET", `/static/.*`, handleStatic),

    newAPIRoute("POST", `/api/v01/history`, NewHistoryRequest, handleHistory),
    newAPIRoute("POST", `/api/v01/peek`, NewPeekRequest, handlePeek),
    newAPIRoute("POST", `/api/v01/submit`, NewSubmissionRequest, handleSubmit),
}

// Handlers that internally handle and log errors should return nil and ensure that responses are written.
type RouteHandler func(response http.ResponseWriter, request *http.Request) error;

// A handler specifically for API endpoints.
// This type alias is not actually used (since type parameters cannot be used in func types), but is supplied here for reference.
// The int will be used as the status code.
// If the status is zero, then a default status 200 or 500 will be used.
// If the status is negative, then nothing will be written to the response and no status will be set
// (it will be assumed that the handler already handled the response).
// The any object will be json encoded and passed back in the response.
// If the error is not nil, then it will be logged and an error status will be set (if not already specified).
type APIHandler func(model.APIRequest) (int, any, error);

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

func newAPIRoute[T model.APIRequest](method string, pattern string,
        requestFactory func(*http.Request) (T, *model.APIResponse, error),
        apiHandler func(T) (int, any, error)) route {
    handler := func(response http.ResponseWriter, request *http.Request) error {
        return handleAPIEndpoint(response, request, requestFactory, apiHandler);
    }

    return route{method, regexp.MustCompile("^" + pattern + "$"), handler};
}

func newRedirect(method string, pattern string, target string) route {
    redirectFunc := func(response http.ResponseWriter, request *http.Request) error {
        return handleRedirect(target, response, request);
    };
    return route{method, regexp.MustCompile("^" + pattern + "$"), redirectFunc};
}

func handleRedirect(target string, response http.ResponseWriter, request *http.Request) error {
    http.Redirect(response, request, target, 301);
    return nil;
}

func handleAPIEndpoint[T model.APIRequest](response http.ResponseWriter, request *http.Request,
        requestFactory func(*http.Request) (T, *model.APIResponse, error),
        apiHandler func(T) (int, any, error)) error {
    apiRequest, apiResponse, err := requestFactory(request);
    if (err != nil) {
        log.Info().Err(err).Msg("Could not deserialize API request.");
        http.Error(response, "Could not deserialize API request.", http.StatusBadRequest);
        return nil;
    } else if (apiResponse != nil) {
        // Short-cut the response.
        sendAPIResponse(response, request, apiResponse);
        return nil;
    }
    defer apiRequest.Close();

    status, message, err := apiHandler(apiRequest);
    if (err != nil) {
        log.Error().Err(err).Str("path", request.URL.Path).Msg("Handler had an error.");
    }

    if (status < 0) {
        return nil;
    }

    if (status == 0) {
        if (err != nil) {
            status = http.StatusInternalServerError;
        } else {
            status = http.StatusOK;
        }
    }

    apiResponse = model.NewResponse(status, message);
    sendAPIResponse(response, request, apiResponse);

    return nil;
}

func sendAPIResponse(response http.ResponseWriter, request *http.Request, apiResponse *model.APIResponse) {
    err := apiResponse.Send(response);
    if (err != nil) {
        log.Error().Err(err).Str("path", request.URL.Path).Msg("Error sending API response.");
        http.Error(response, "Server Error", http.StatusInternalServerError);
    }
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
