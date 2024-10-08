package upsert

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

type commonTestCase struct {
	email                string
	path                 string
	expectedLocator      string
	expectedCount        int
	expectedSuccessCount int
}

func processRsponse(test *testing.T, response *core.APIResponse, testCase commonTestCase, prefix string) {
	if !response.Success {
		if testCase.expectedLocator != "" {
			if testCase.expectedLocator != response.Locator {
				test.Errorf("%sIncorrect error returned. Expected '%s', found '%s'.",
					prefix, testCase.expectedLocator, response.Locator)
			}
		} else {
			test.Errorf("%sResponse is not a success when it should be: '%v'.", prefix, response)
		}

		return
	}

	if testCase.expectedLocator != "" {
		test.Errorf("%sDid not get an expected error.", prefix)
		return
	}

	var responseContent UpsertResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	actualCount := len(responseContent.Results)
	if testCase.expectedCount != actualCount {
		test.Errorf("%sUnexpected course count. Expected: %d, actual: %d.",
			prefix, testCase.expectedCount, actualCount)
		return
	}

	actualSuccessCount := 0
	for _, result := range responseContent.Results {
		if result.Success {
			actualSuccessCount++
		}
	}

	if testCase.expectedSuccessCount != actualSuccessCount {
		test.Errorf("%sUnexpected successful course count. Expected: %d, actual: %d.",
			prefix, testCase.expectedSuccessCount, actualSuccessCount)
		return
	}
}
