package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/users"
	"github.com/edulinq/autograder/internal/util"
)

func TestUpsert(test *testing.T) {
	testCases := []struct {
		email     string
		permError bool
		options   *users.UpsertUsersOptions
		expected  []*model.UserOpResult
	}{
		// Valid permissions.
		{"server-admin", false, &users.UpsertUsersOptions{}, []*model.UserOpResult{}},
		{"server-owner", false, &users.UpsertUsersOptions{}, []*model.UserOpResult{}},

		// Invalid permissions.
		{"server-user", true, nil, nil},
		{"server-creator", true, nil, nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"options": testCase.options,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`users/upsert`), fields, nil, testCase.email)
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

		var responseContent UpsertResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(testCase.expected, responseContent.Results) {
			test.Errorf("Case %d: Unexpected user op result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent.Results))
			continue
		}
	}
}
