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
		expected bool
	}{
		// Test cases for correct authorization.
		{"other@test.com", "other", true},
		{"student@test.com", "student", true},
		{"grader@test.com", "grader", true},
		{"admin@test.com", "admin", true},
		{"owner@test.com", "owner", true},

		// Ensure we fail on bad passwords.
		{"other@test.com", "ZZZ", false},
		{"student@test.com", "ZZZ", false},
		{"grader@test.com", "ZZZ", false},
		{"admin@test.com", "ZZZ", false},
		{"owner@test.com", "ZZZ", false},

		// Check we cannot find invalid users.
		{"ZZZ", "Z", false},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"user-email": testCase.email,
			"user-pass":  util.Sha256HexFromString(testCase.pass),
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint("users/auth"), fields)
		if testCase.expected != response.Success {
			test.Errorf("Case %d: Unexpected result. Expected: '%t', actual: '%t'.", i, testCase.expected, response.Success)
			continue
		}

		var responseContent AuthResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if testCase.expected != responseContent.Success {
			test.Errorf("Case %d: Unexpected result. Expected: '%t', actual: '%t'.", i, testCase.expected, responseContent.Success)
			continue
		}
	}
}
