package tokens

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestCreateBase(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		email       string
		targetUser  string
		name        string
		missingUser bool
		locator     string
	}{
		// Self
		{
			email: "course-student@test.edulinq.org",
		},

		// Name
		{
			email: "course-student@test.edulinq.org",
			name:  "test",
		},

		// Other
		{
			email:      "server-admin@test.edulinq.org",
			targetUser: "course-student@test.edulinq.org",
		},

		// Other - Bad Permissions
		{
			email:      "course-admin@test.edulinq.org",
			targetUser: "course-student@test.edulinq.org",
			locator:    "-046",
		},

		// Other - Missing
		{
			email:       "server-admin@test.edulinq.org",
			targetUser:  "ZZZ",
			missingUser: true,
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		targetEmail := testCase.targetUser
		if targetEmail == "" {
			targetEmail = testCase.email
		}

		targetUser, err := db.GetServerUser(targetEmail)
		if err != nil {
			test.Errorf("Case %d: Failed to get initial target user: '%v'.", i, err)
			continue
		}

		initialTokenCount := 0
		if targetUser != nil {
			initialTokenCount = len(targetUser.Tokens)
		}

		fields := map[string]any{
			"target-user": testCase.targetUser,
			"name":        testCase.name,
		}

		response := core.SendTestAPIRequestFull(test, "users/tokens/create", fields, nil, testCase.email)
		if !response.Success {
			if testCase.locator != response.Locator {
				test.Errorf("Case %d: Incorrect error returned. Expected: '%s', Actual: '%s'.",
					i, testCase.locator, response.Locator)
			}

			continue
		}

		if testCase.locator != "" {
			test.Errorf("Case %d: Did not get an expected error. Expected: '%s'", i, testCase.locator)
			continue
		}

		var responseContent CreateResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		targetUser, err = db.GetServerUser(targetEmail)
		if err != nil {
			test.Errorf("Case %d: Failed to get final target user: '%v'.", i, err)
			continue
		}

		if targetUser == nil {
			if !testCase.missingUser {
				test.Errorf("Case %d: Found no user when one was expected.", i)
				continue
			}

			if responseContent.FoundUser {
				test.Errorf("Case %d: Found user not as expected. Expected: false, Actual: true.", i)
				continue
			}

			continue
		}

		if testCase.missingUser {
			test.Errorf("Case %d: Found a user when none was expected.", i)
			continue
		}

		if !responseContent.FoundUser {
			test.Errorf("Case %d: Found user not as expected. Expected: true, Actual: false.", i)
			continue
		}

		finalTokenCount := len(targetUser.Tokens)

		if finalTokenCount != (initialTokenCount + 1) {
			test.Errorf("Case %d: Unexpected token count. Expected: %d, Actual: %d.", i, (initialTokenCount + 1), finalTokenCount)
			continue
		}

		// Get the newest token.
		var newToken *model.Token
		for _, token := range targetUser.Tokens {
			if (newToken == nil) || (newToken.CreationTime < token.CreationTime) {
				newToken = token
			}
		}

		if !reflect.DeepEqual(newToken.TokenInfo, *responseContent.TokenInfo) {
			test.Errorf("Case %d: New token info not as expected. Expected: %s, Actual: %s.",
				i, util.MustToJSONIndent(newToken.TokenInfo), util.MustToJSONIndent(responseContent.TokenInfo))
			continue
		}

		if testCase.name != newToken.Name {
			test.Errorf("Case %d: New token name not as expected. Expected: '%s', Actual: '%s'.",
				i, testCase.name, newToken.Name)
			continue
		}
	}
}
