package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestPassChange(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		newPass  string
		expected PasswordChangeResponse
	}{
		{"spooky", PasswordChangeResponse{true, false}},
		{"admin", PasswordChangeResponse{true, true}},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		fields := map[string]any{
			"user-email": "course-admin@test.edulinq.org",
			"user-pass":  util.Sha256HexFromString("admin"),
			"new-pass":   util.Sha256HexFromString(testCase.newPass),
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint(`users/password/change`), fields)
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			continue
		}

		var responseContent PasswordChangeResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(testCase.expected, responseContent) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}

		user, err := db.GetServerUser("course-admin@test.edulinq.org", true)
		if err != nil {
			test.Errorf("Case %d: Failed to get saved user: '%v'.", i, err)
			continue
		}

		success, err := user.Auth(util.Sha256HexFromString(testCase.newPass))
		if err != nil {
			test.Errorf("Case %d: Failed to auth user: '%v'.", i, err)
			continue
		}

		if !success {
			test.Errorf("Case %d: The new password fails to auth after the change.", i)
			continue
		}
	}
}
