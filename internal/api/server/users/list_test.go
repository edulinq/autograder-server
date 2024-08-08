package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestList(test *testing.T) {
	usersMap := db.MustGetServerUsers()

	users := make([]*model.ServerUser, 0, len(usersMap))
	for _, user := range usersMap {
		users = append(users, user)
	}

	testCases := []struct {
		email     string
		permError bool
	}{
		// Invalid permissions.
		{"server-user@test.com", true},
		{"server-creator@test.com", true},

		// Valid permissions.
		{"server-admin@test.com", false},
		{"server-owner@test.com", false},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"user-email": testCase.email,
			"user-pass":  util.Sha256HexFromString(*usersMap[testCase.email].Name),
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint(`server/users/list`), fields)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-041"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
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

		var responseContent ListResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		expectedInfos := core.MustNewServerUserInfos(users)
		if !reflect.DeepEqual(expectedInfos, responseContent.Users) {
			test.Errorf("Case %d: Unexpected users information. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(expectedInfos), util.MustToJSONIndent(responseContent.Users))
			continue
		}
	}
}
