package api

import (
    "fmt"
    "math"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

// The most simple authenticating request.
type BaseTestRequest struct {
    APIRequestCourseUserContext
    MinRoleStudent
}

// Force a panic from an API handler.
func TestAPIPanic(test *testing.T) {
    endpoint := `/test/api/panic`;

    handler := func(request *BaseTestRequest) (*any, *APIError) {
        panic("Forced Panic!");
        return nil, nil;
    }

    routes = append(routes, newAPIRoute(endpoint, handler));

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    response := sendTestAPIRequest(test, endpoint, nil);
    if (response.Locator != "-501") {
        test.Fatalf("Response does not have panic locator of '-501', actual locator: '%s'.", response.Locator);
    }
}

// Test handlers that do not have the correct signature.
// Specifically, we are focusing on testing validateAPIHandler().
func TestMalformedHandlers(test *testing.T) {
    // Define all the handlers.
    testCases := []struct{handler any; locator string}{
        {"", "-521"},
        {nil, "-521"},
        {0, "-521"},
        {func() (*any, *APIError) { return nil, nil }, "-522"},
        {func(request *BaseTestRequest, testarg int) (*any, *APIError) { return nil, nil }, "-522"},
        {func(request BaseTestRequest) (*any, *APIError) { return nil, nil }, "-523"},
        {func(request int) (*any, *APIError) { return nil, nil }, "-523"},
        {func(request *BaseTestRequest) (*any) { return nil }, "-524"},
        {func(request *BaseTestRequest) (int, *any, *APIError) { return 0, nil, nil }, "-524"},
        {func(request *BaseTestRequest) (any, *APIError) { return nil, nil }, "-525"},
        {func(request *BaseTestRequest) (int, *APIError) { return 0, nil }, "-525"},
        {func(request *BaseTestRequest) (*any, APIError) { return nil, APIError{} }, "-526"},
        {func(request *BaseTestRequest) (*any, any) { return nil, nil }, "-526"},
        {func(request *BaseTestRequest) (*any, int) { return nil, 0}, "-526"},
    };

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    for i, testCase := range testCases {
        // Register the handlers using its index in the endpoint..
        endpoint := fmt.Sprintf("/test/api/malformed/handler/%d", i);
        routes = append(routes, newAPIRoute(endpoint, testCase.handler));

        response := sendTestAPIRequest(test, endpoint, nil);
        if (response.Locator != testCase.locator) {
            test.Errorf("Case %d -- Expected response locator of '%s', found response locator of '%s'. Response: [%v]", i, testCase.locator, response.Locator, response);
        }
    }
}

// Test empty/non-deserializable content.
func TestBadRequestEmptyContent(test *testing.T) {
    // Define all the content that will go in the post form.
    testCases := []struct{form map[string]string; locator string}{
        {map[string]string{}, "-402"},
        {map[string]string{API_REQUEST_CONTENT_KEY: ``}, "-402"},
        {map[string]string{API_REQUEST_CONTENT_KEY: `Z`}, "-403"},
        {map[string]string{API_REQUEST_CONTENT_KEY: `1`}, "-403"},
        {map[string]string{API_REQUEST_CONTENT_KEY: `[]`}, "-403"},
    };

    endpoint := `/test/api/bad-request/empty-content`;
    handler := func(request *BaseTestRequest) (*any, *APIError) { return nil, nil };
    routes = append(routes, newAPIRoute(endpoint, handler));

    url := serverURL + endpoint;

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    for i, testCase := range testCases {
        responseText, err := util.PostNoCheck(url, testCase.form);
        if (err != nil) {
            test.Errorf("Case %d -- POST returned an error: '%v'.", i, err);
            continue;
        }

        var response APIResponse;
        err = util.JSONFromString(responseText, &response);
        if (err != nil) {
            test.Errorf("Case %d -- Could not unmarshal JSON response '%s': '%v'.", i, responseText, err);
            continue;
        }

        if (response.Locator != testCase.locator) {
            test.Errorf("Case %d -- Expected response locator of '%s', found response locator of '%s'. Response: [%v]", i, testCase.locator, response.Locator, response);
        }
    }
}

// Return an object from the handler than cannot be marshaled.
func TestNonMarshalableResponse(test *testing.T) {
    endpoint := `/test/api/bad-response/non-marshalable`;

    type responseType struct {
        Value float64
    }

    handler := func(request *BaseTestRequest) (*responseType, *APIError) {
        response := responseType{
            Value: math.NaN(),
        };

        return &response, nil;
    }

    routes = append(routes, newAPIRoute(endpoint, handler));

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    response := sendTestAPIRequest(test, endpoint, nil);
    if (response.Locator != "-531") {
        test.Fatalf("Response does not locator of '-531', actual locator: '%s'.", response.Locator);
    }
}
