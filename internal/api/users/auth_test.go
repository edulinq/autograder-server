package users

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestUsersAuth(test *testing.T) {
	testCases := []struct {
		email    string
		pass     string
		expected AuthResponse
	}{
		// Test cases for correct authorization.
		{"other@test.com", "other", AuthResponse{true, true}},
		{"student@test.com", "student", AuthResponse{true, true}},
		{"grader@test.com", "grader", AuthResponse{true, true}},
		{"admin@test.com", "admin", AuthResponse{true, true}},
		{"owner@test.com", "owner", AuthResponse{true, true}},

		// Ensure we fail on bad passwords.
		{"other@test.com", "ZZZ", AuthResponse{true, false}},
		{"student@test.com", "ZZZ", AuthResponse{true, false}},
		{"grader@test.com", "ZZZ", AuthResponse{true, false}},
		{"admin@test.com", "ZZZ", AuthResponse{true, false}},
		{"owner@test.com", "ZZZ", AuthResponse{true, false}},

		// Check we cannot find invalid users.
		{"ZZZ", "Z", AuthResponse{false, false}},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-email": testCase.email,
			"target-pass":  util.Sha256HexFromString(testCase.pass),
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint("users/auth"), fields)
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			continue
		}

		var responseContent AuthResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if testCase.expected != responseContent {
			test.Errorf("Case %d: Unexpected result. Expected: '%+v', actual: '%+v'.", i, testCase.expected, responseContent)
			continue
		}
	}
}
