package assignments

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestGet(test *testing.T) {
	users := db.MustGetServerUsers()

	assignment := db.MustGetTestAssignment()
	expected := core.NewAssignmentInfo(assignment)

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
		{"admin@test.com", false},
		{"other@test.com", false},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"user-email": testCase.email,
			"user-pass":  util.Sha256HexFromString(*users[testCase.email].Name),
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint(`courses/assignments/get`), fields)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-040"
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

		var responseContent GetResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(expected, responseContent.Assignment) {
			test.Fatalf("Unexpected result. Expected: '%s', actual: '%s'.",
				util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent.Assignment))
			continue
		}
	}
}
