package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestUserRemove(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		email     string
		target    string
		permError bool
		locator   string
		expected  RemoveResponse
	}{
		// Valid permissions.
		{"server-owner", "server-user@test.edulinq.org", false, "", RemoveResponse{true}},
		{"server-owner", "server-creator@test.edulinq.org", false, "", RemoveResponse{true}},
		{"server-owner", "server-admin@test.edulinq.org", false, "", RemoveResponse{true}},

		{"server-owner", "server-user@test.edulinq.org", false, "", RemoveResponse{true}},
		{"server-admin", "server-user@test.edulinq.org", false, "", RemoveResponse{true}},

		// Invalid permissions.
		{"server-creator", "server-user@test.edulinq.org", true, "-041", RemoveResponse{false}},
		{"server-user", "server-user@test.edulinq.org", true, "-041", RemoveResponse{false}},

		// Target user is not found.
		{"server-owner", "ZZZ", false, "", RemoveResponse{false}},

		// Complex invalid permissions.
		{"server-admin", "server-owner@test.edulinq.org", true, "-811", RemoveResponse{false}},
		{"server-owner", "server-owner@test.edulinq.org", true, "-811", RemoveResponse{false}},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		initialUser, err := db.GetServerUser(testCase.target)
		if err != nil {
			test.Errorf("Case %d: Failed to get initial user: '%v'.", i, err)
			continue
		}

		fields := map[string]any{
			"target-email": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`users/remove`), fields, nil, testCase.email)

		if !response.Success {
			if testCase.permError {
				if testCase.locator != response.Locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.locator, response.Locator)
					continue
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
				continue
			}
		}

		var responseContent RemoveResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if testCase.expected != responseContent {
			test.Errorf("Case %d: Unexpected remove result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}

		user, err := db.GetServerUser(testCase.target)
		if err != nil {
			test.Errorf("Case %d: Failed to get user: '%v'.", i, err)
			continue
		}

		if !testCase.expected.FoundUser {
			// Ensure that the user is NOT removed.
			if !reflect.DeepEqual(initialUser, user) {
				test.Errorf("Case %d: Unexpected user change. Expected: '%+v', actual: '%+v'.",
					i, initialUser, user)
				continue
			}
		} else {
			// Ensure that the user is removed.
			if user != nil {
				test.Errorf("Case %d: User not removed from DB: '%+v'.", i, user)
				continue
			}
		}
	}
}
