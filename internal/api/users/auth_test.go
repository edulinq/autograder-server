package users

import (
	// "fmt"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestUsersAuth(test *testing.T) {
	testCases := []struct {
		role     model.ServerUserRole
		email    string
		pass     string
		expected AuthResponse
	}{
		// Test cases for correct authorization.
		{model.ServerRoleOwner, "owner@test.com", "owner", AuthResponse{true, true}},
		{model.ServerRoleUser, "student@test.com", "student", AuthResponse{true, true}},

		// Ensure we fail on bad passwords.
		{model.ServerRoleUser, "student@test.com", "ZZZ", AuthResponse{true, false}},

		// Check we cannot find invalid users.
		{model.ServerRoleUser, "ZZZ", "Z", AuthResponse{false, false}},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-email": testCase.email,
			"target-pass":  util.Sha256HexFromString(testCase.pass),
		}

		// attemptPass := fields["target-pass"].(string);
		// fmt.Printf("Trying to pass: '%s'.\n", attemptPass);

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint("users/auth"), fields, nil, model.CourseRoleUnknown)
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
