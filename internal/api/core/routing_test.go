package core

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"
)

// Define all the malformed handlers.
var handlerTestCases = []malformedHandlerTestCase{
	{"", "-006"},
	{nil, "-006"},
	{0, "-006"},
	{func() (*any, *APIError) { return nil, nil }, "-007"},
	{func(request *BaseTestRequest, testarg int) (*any, *APIError) { return nil, nil }, "-007"},
	{func(request BaseTestRequest) (*any, *APIError) { return nil, nil }, "-008"},
	{func(request int) (*any, *APIError) { return nil, nil }, "-008"},
	{func(request *BaseTestRequest) *any { return nil }, "-009"},
	{func(request *BaseTestRequest) (int, *any, *APIError) { return 0, nil, nil }, "-009"},
	{func(request *BaseTestRequest) (any, *APIError) { return nil, nil }, "-010"},
	{func(request *BaseTestRequest) (int, *APIError) { return 0, nil }, "-010"},
	{func(request *BaseTestRequest) (*any, APIError) { return nil, APIError{} }, "-011"},
	{func(request *BaseTestRequest) (*any, any) { return nil, nil }, "-011"},
	{func(request *BaseTestRequest) (*any, int) { return nil, 0 }, "-011"},
}

// The most simple authenticating request.
type BaseTestRequest struct {
	APIRequestCourseUserContext
	MinCourseRoleStudent
}

type malformedHandlerTestCase struct {
	handler any
	locator string
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
	// Check to see if we are trying to crash.
	if os.Getenv("BE_CRASHER") == "1" {
		index, err := strconv.Atoi(os.Getenv("TEST_CASE_INDEX"))
		if err != nil {
			test.Fatalf("Failed to parse TEST_CASE_INDEX: '%v'.", err)
		}

		// Register the handlers using its index in the endpoint.
		endpoint := fmt.Sprintf("/test/api/malformed/handler/%d", index)
		handler := handlerTestCases[index].handler

		// We expect this to fail HARD.
		MustNewAPIRoute(endpoint, handler)

		return
	}

	errorPrefix := "API Error -- "
	re := regexp.MustCompile(errorPrefix + `\{.*\}`)

	for i, testCase := range handlerTestCases {
		cmd := exec.Command(os.Args[0], "-test.run=TestMalformedHandlers")
		cmd.Env = append(os.Environ(), "BE_CRASHER=1", fmt.Sprintf("TEST_CASE_INDEX=%d", i))

		var buffer bytes.Buffer
		cmd.Stdout = &buffer
		cmd.Stderr = &buffer

		err := cmd.Run()
		output := buffer.String()

		// Verify the process exited with the correct error.
		exitErr, ok := err.(*exec.ExitError)
		if ok && !exitErr.Success() {
			if !strings.Contains(output, "Error while validating API handler.") {
				test.Errorf("Case %d: Unexpected fatal error message: '%s'.", i, output)
				continue
			}

			jsonWithPrefix := re.FindString(output)
			if jsonWithPrefix == "" {
				test.Errorf("Case %d: Unable to find JSON error message: '%s'.", i, output)
				continue
			}

			jsonOutput := strings.TrimPrefix(jsonWithPrefix, errorPrefix)

			var apiError APIError
			util.MustJSONFromString(jsonOutput, &apiError)

			endpoint := MakeFullAPIPath(fmt.Sprintf("/test/api/malformed/handler/%d", i))
			if endpoint != apiError.Endpoint {
				test.Errorf("Case %d: Unxpected endpoint. Expected: '%s', actual: '%s'.",
					i, endpoint, apiError.Endpoint)
				continue
			}

			if testCase.locator != apiError.Locator {
				test.Errorf("Case %d: Unxpected locator. Expected: '%s', actual: '%s'.",
					i, testCase.locator, apiError.Locator)
				continue
			}

			continue
		}

		test.Errorf("Case %d: Expected process to fail, but it succeeded. Output: '%s'.", i, output)
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
		responseText, err := common.PostNoCheck(url, testCase.form)
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
