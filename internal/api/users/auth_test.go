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
		{"course-other@test.edulinq.org", "course-other", true},
		{"course-student@test.edulinq.org", "course-student", true},
		{"course-grader@test.edulinq.org", "course-grader", true},
		{"course-admin@test.edulinq.org", "course-admin", true},
		{"course-owner@test.edulinq.org", "course-owner", true},

		// Ensure we fail on bad passwords.
		{"course-other@test.edulinq.org", "ZZZ", false},
		{"course-student@test.edulinq.org", "ZZZ", false},
		{"course-grader@test.edulinq.org", "ZZZ", false},
		{"course-admin@test.edulinq.org", "ZZZ", false},
		{"course-owner@test.edulinq.org", "ZZZ", false},

		// Check we cannot find invalid users.
		{"ZZZ", "Z", false},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"user-email": testCase.email,
			"user-pass":  util.Sha256HexFromString(testCase.pass),
		}

		response := core.SendTestAPIRequest(test, `users/auth`, fields)
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
