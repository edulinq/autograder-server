package core

import (
	"fmt"
	"math"
	"testing"

	"github.com/edulinq/autograder/internal/exit"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

// The most simple authenticating request.
type BaseTestRequest struct {
	APIRequestCourseUserContext
	MinCourseRoleStudent
}

// Force a panic from an API handler.
func TestAPIPanic(test *testing.T) {
	endpoint := `/test/api/panic`

	handler := func(request *BaseTestRequest) (*any, *APIError) {
		panic("Forced Panic!")
		return nil, nil
	}

	routes = append(routes, MustNewAPIRoute(endpoint, handler))

	response := SendTestAPIRequest(test, endpoint, nil)
	if response.Locator != "-001" {
		test.Fatalf("Response does not have panic locator of '-001', actual locator: '%s'.", response.Locator)
	}
}

// Test handlers that do not have the correct signature.
// Specifically, we are focusing on testing validateAPIHandler().
func TestMalformedHandlers(test *testing.T) {
	// Suppress exits to capture exit codes.
	exit.SetShouldExitForTesting(false)
	defer exit.SetShouldExitForTesting(true)

	oldLogLevel := log.GetTextLevel()
	log.SetTextLevel(log.LevelOff)
	defer log.SetTextLevel(oldLogLevel)

	// Define all the malformed handlers.
	testCases := []struct {
		handler any
	}{
		{""},
		{nil},
		{0},
		{func() (*any, *APIError) { return nil, nil }},
		{func(request *BaseTestRequest, testarg int) (*any, *APIError) { return nil, nil }},
		{func(request BaseTestRequest) (*any, *APIError) { return nil, nil }},
		{func(request int) (*any, *APIError) { return nil, nil }},
		{func(request *BaseTestRequest) *any { return nil }},
		{func(request *BaseTestRequest) (int, *any, *APIError) { return 0, nil, nil }},
		{func(request *BaseTestRequest) (any, *APIError) { return nil, nil }},
		{func(request *BaseTestRequest) (int, *APIError) { return 0, nil }},
		{func(request *BaseTestRequest) (*any, APIError) { return nil, APIError{} }},
		{func(request *BaseTestRequest) (*any, any) { return nil, nil }},
		{func(request *BaseTestRequest) (*any, int) { return nil, 0 }},
	}

	for i, testCase := range testCases {
		// Register the handlers using its index in the endpoint.
		endpoint := fmt.Sprintf("/test/api/malformed/handler/%d", i)

		MustNewAPIRoute(endpoint, testCase.handler)

		// Verify the process exited with the correct error.
		exitCode := exit.GetLastExitCode()
		if exitCode != exit.EXIT_SOFTWARE {
			test.Errorf("Case %d: Unexpected exit code. Expected: '%d', actual: '%d'.", i, exit.EXIT_SOFTWARE, exitCode)
		}
	}
}

// Test empty/non-deserializable content.
func TestBadRequestEmptyContent(test *testing.T) {
	// Define all the content that will go in the post form.
	testCases := []struct {
		form    map[string]string
		locator string
	}{
		{map[string]string{}, "-004"},
		{map[string]string{API_REQUEST_CONTENT_KEY: ``}, "-004"},
		{map[string]string{API_REQUEST_CONTENT_KEY: `Z`}, "-005"},
		{map[string]string{API_REQUEST_CONTENT_KEY: `1`}, "-005"},
		{map[string]string{API_REQUEST_CONTENT_KEY: `[]`}, "-005"},
	}

	endpoint := `/test/api/bad-request/empty-content`
	handler := func(request *BaseTestRequest) (*any, *APIError) { return nil, nil }
	routes = append(routes, MustNewAPIRoute(endpoint, handler))

	url := serverURL + MakeFullAPIPath(endpoint)

	for i, testCase := range testCases {
		responseText, err := util.PostNoCheck(url, testCase.form)
		if err != nil {
			test.Errorf("Case %d: POST returned an error: '%v'.", i, err)
			continue
		}

		var response APIResponse
		err = util.JSONFromString(responseText, &response)
		if err != nil {
			test.Errorf("Case %d: Could not unmarshal JSON response '%s': '%v'.", i, responseText, err)
			continue
		}

		if response.Locator != testCase.locator {
			test.Errorf("Case %d: Expected response locator of '%s', found response locator of '%s'. Response: [%v]", i, testCase.locator, response.Locator, response)
		}
	}
}

// Return an object from the handler than cannot be marshaled.
func TestNonMarshalableResponse(test *testing.T) {
	endpoint := `/test/api/bad-response/non-marshalable`

	type responseType struct {
		Value float64
	}

	handler := func(request *BaseTestRequest) (*responseType, *APIError) {
		response := responseType{
			Value: math.NaN(),
		}

		return &response, nil
	}

	routes = append(routes, MustNewAPIRoute(endpoint, handler))

	response := SendTestAPIRequest(test, endpoint, nil)
	if response.Locator != "-002" {
		test.Fatalf("Response does not locator of '-002', actual locator: '%s'.", response.Locator)
	}
}
