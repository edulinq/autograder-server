package tokens

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestListBase(test *testing.T) {
	testCases := []struct {
		email       string
		targetUser  string
		missingUser bool
		locator     string
	}{
		// Self
		{
			email: "course-student@test.edulinq.org",
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
		fields := map[string]any{
			"target-user": testCase.targetUser,
		}

		response := core.SendTestAPIRequestFull(test, "users/tokens/list", fields, nil, testCase.email)
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

		var responseContent ListResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		targetEmail := testCase.targetUser
		if targetEmail == "" {
			targetEmail = testCase.email
		}

		expectedUser, err := db.GetServerUser(targetEmail)
		if err != nil {
			test.Errorf("Case %d: Failed to get expected user: '%v'.", i, err)
			continue
		}

		if expectedUser == nil {
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

		expectedTokens := make([]model.TokenInfo, 0, len(expectedUser.Tokens))
		for _, token := range expectedUser.Tokens {
			expectedTokens = append(expectedTokens, token.TokenInfo)
		}

		if len(expectedTokens) == 0 {
			test.Errorf("Case %d: Found no expected tokens.", i)
			continue
		}

		if !reflect.DeepEqual(expectedTokens, responseContent.Tokens) {
			test.Errorf("Unexpected tokens. Expected: %s, Actual: %s.",
				util.MustToJSONIndent(expectedTokens), util.MustToJSONIndent(responseContent.Tokens))
		}
	}
}
