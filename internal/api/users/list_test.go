package users

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestList(test *testing.T) {
	users := db.MustGetServerUsers()

	testCases := []struct {
		contextUser *model.ServerUser
		permError   bool
		expected    []*core.ServerUserInfo
	}{
		// Bad permissions.
		{users["server-user@test.com"], true, nil},

		// Good permissions.
		// TODO: Add expected output.
		{users["server-admin@test.com"], false, nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"user-email": testCase.contextUser.Email,
			"user-pass":  util.Sha256HexFromString(*testCase.contextUser.Name),
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint(`users/list`), fields)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-041"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.", i, expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.permError {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		var responseContent GetResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		expectedFound := (testCase.expected != nil)
		if (expectedFound != responseContent.Found) {
			test.Errorf("Case %d: Server user info does not match. Expected: '%v', actual: '%v'.", i, expectedFound, responseContent.Found)
			continue
		}

		test.Errorf("Case %d: Found the following user info: '%s'.", i, util.MustToJSON(responseContent))
		if testCase.expected == nil {
			continue
		}

		// TODO: Add check for actual content!
	}
}
