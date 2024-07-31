package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestGet(test *testing.T) {
	users := db.MustGetServerUsers()

	testCases := []struct {
		contextUser *model.ServerUser
		target      string
		permError   bool
		expected    *model.ServerUser
	}{
		// Self, empty.
		{users["student@test.com"], "", false, users["student@test.com"]},
		{users["server-creator@test.com"], "", false, users["server-creator@test.com"]},
		{users["server-admin@test.com"], "", false, users["server-admin@test.com"]},

		// Other, bad permissions.
		{users["server-creator@test.com"], "student@test.com", true, nil},

		// Other, good permissions.
		{users["server-admin@test.com"], "student@test.com", false, users["student@test.com"]},

		// Missing
		{users["server-creator@test.com"], "ZZZ", true, nil},
		{users["server-admin@test.com"], "ZZZ", false, nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"user-email":   testCase.contextUser.Email,
			"user-pass":    util.Sha256HexFromString(*testCase.contextUser.Name),
			"target-email": testCase.target,
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint(`server/users/get`), fields)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-046"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expcted '%s', found '%s'.",
						i, expectedLocator, response.Locator)
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
		if expectedFound != responseContent.Found {
			test.Errorf("Case %d: Found user does not match. Expected: '%v', actual: '%v'.", i, expectedFound, responseContent.Found)
			continue
		}

		if testCase.expected == nil {
			continue
		}

		expectedInfo := core.MustNewServerUserInfo(testCase.expected)
		if !reflect.DeepEqual(expectedInfo, responseContent.User) {
			test.Errorf("Case %d: Unexpected user result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(expectedInfo), util.MustToJSONIndent(responseContent.User))
			continue
		}
	}
}
