package api

import (
    "fmt"
    "testing"

    "github.com/eriq-augustine/autograder/config"
)

// The most simple authenticating request.
type BaseTestRequest struct {
    APIRequestCourseUserContext
    MinRoleStudent
}

// Force a panic from an API handler.
func TestAPIPanic(test *testing.T) {
    panicEndpoint := `/test/api/panic`;

    panicHandler := func(request *BaseTestRequest) (*any, *APIError) {
        panic("Forced Panic!");
        return nil, nil;
    }

    routes = append(routes, newAPIRoute(panicEndpoint, panicHandler));

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    response := sendTestAPIRequest(test, panicEndpoint, nil);
    if (response.ID != "-501") {
        test.Fatalf("Response does not have panic ID of '-501', actual ID: '%s'.", response.ID);
    }
}

// Test handlers that do not have the correct signature.
// Specifically, we are focusing on testing validateAPIHandler().
func TestMalformedHandlers(test *testing.T) {
    // Define all the handlers.
    testCases := []struct{handler any; id string}{
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
        if (response.ID != testCase.id) {
            test.Errorf("Case %d -- Expected response ID of '%s', found response ID of '%s'. Response: [%v]", i, testCase.id, response.ID, response);
        }
    }
}
