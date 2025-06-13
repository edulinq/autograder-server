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
		if user.Email == model.RootUserEmail {
			continue
		}

		users = append(users, user)
	}

	testCases := []struct {
		email         string
		input         []string
		expectedUsers []*model.ServerUser
		locator       string
	}{
		// Invalid Permissions
		{"server-user", nil, nil, "-041"},
		{"server-creator", nil, nil, "-041"},

		// Valid Permissions, All Users
		{"server-admin", nil, users, ""},
		{"server-owner", []string{}, users, ""},
		{"server-owner", []string{"*"}, users, ""},

		// No Users
		{"server-admin", []string{"-*"}, []*model.ServerUser{}, ""},

		// One User
		{"server-admin", []string{"owner"}, []*model.ServerUser{usersMap["server-owner@test.edulinq.org"]}, ""},

		// Input Errors
		{"server-admin", []string{"ZZZ"}, nil, "-815"},
		{"server-admin", []string{"ZZZ::*"}, nil, "-815"},
		{"server-admin", []string{"*::ZZZ"}, nil, "-815"},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-users": testCase.input,
		}

		response := core.SendTestAPIRequestFull(test, `users/list`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.locator != "" {
				if response.Locator != testCase.locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.locator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.locator != "" {
			test.Errorf("Case %d: Did not get an expected error: '%s'.", i, testCase.locator)
			continue
		}

		var responseContent ListResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		expectedInfos := core.NewServerUserInfos(testCase.expectedUsers)

		if !reflect.DeepEqual(expectedInfos, responseContent.Users) {
			test.Errorf("Case %d: Unexpected users information. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(expectedInfos), util.MustToJSONIndent(responseContent.Users))
			continue
		}
	}
}
