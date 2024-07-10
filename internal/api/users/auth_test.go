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
		success  bool
		expected AuthResponse
	}{
		// Test cases for correct authorization.
		{"other@test.com", "other", true, AuthResponse{true}},
		{"student@test.com", "student", true, AuthResponse{true}},
		{"grader@test.com", "grader", true, AuthResponse{true}},
		{"admin@test.com", "admin", true, AuthResponse{true}},
		{"owner@test.com", "owner", true, AuthResponse{true}},

		// Ensure we fail on bad passwords.
		{"other@test.com", "ZZZ", false, AuthResponse{false}},
		{"student@test.com", "ZZZ", false, AuthResponse{false}},
		{"grader@test.com", "ZZZ", false, AuthResponse{false}},
		{"admin@test.com", "ZZZ", false, AuthResponse{false}},
		{"owner@test.com", "ZZZ", false, AuthResponse{false}},

		// Check we cannot find invalid users.
		{"ZZZ", "Z", false, AuthResponse{false}},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"user-email": testCase.email,
			"user-pass":  util.Sha256HexFromString(testCase.pass),
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint("users/auth"), fields)
		if testCase.success != response.Success {
			test.Errorf("Case %d: Unexpected result. Expected: '%t', actual: '%t'.", i, testCase.success, response.Success)
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
