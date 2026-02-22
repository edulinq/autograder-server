package tokens

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestDeleteBase(test *testing.T) {
	defer db.ResetForTesting()

	// See testdata/users.json for tokens.
	testCases := []struct {
		email        string
		targetUser   string
		id           string
		missingUser  bool
		missingToken bool
		locator      string
	}{
		// Self
		{
			email: "course-student@test.edulinq.org",
			id:    "dddbc97c-36e4-43fc-b5a0-478aade61c53",
		},

		// Self - Missing
		{
			email: "course-student@test.edulinq.org",
			id:    "ZZZ",
		},

		// Self - No ID
		{
			email:   "course-student@test.edulinq.org",
			locator: "-038",
		},

		// Other
		{
			email:      "server-admin@test.edulinq.org",
			targetUser: "course-student@test.edulinq.org",
			id:         "dddbc97c-36e4-43fc-b5a0-478aade61c53",
		},

		// Other - Token Missing
		{
			email:      "server-admin@test.edulinq.org",
			targetUser: "course-student@test.edulinq.org",
			id:         "ZZZ",
		},

		// Other - Bad Permissions
		{
			email:      "course-admin@test.edulinq.org",
			targetUser: "course-student@test.edulinq.org",
			locator:    "-046",
		},

		// Other - User Missing
		{
			email:       "server-admin@test.edulinq.org",
			id:          "ZZZ",
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

		var initialToken *model.Token
		if targetUser != nil {
			for _, token := range targetUser.Tokens {
				if token.ID == testCase.id {
					initialToken = token
					break
				}
			}
		}

		fields := map[string]any{
			"target-user": testCase.targetUser,
			"token-id":    testCase.id,
		}

		response := core.SendTestAPIRequestFull(test, "users/tokens/delete", fields, nil, testCase.email)
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

		var responseContent DeleteResponse
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

		if !responseContent.FoundToken {
			if initialToken != nil {
				test.Errorf("Case %d: Found token when none was expected.", i)
				continue
			}

			continue
		}

		if initialToken == nil {
			test.Errorf("Case %d: Found no token when one was expected.", i)
			continue
		}

		var finalToken *model.Token
		for _, token := range targetUser.Tokens {
			if token.ID == testCase.id {
				finalToken = token
				break
			}
		}

		if finalToken != nil {
			test.Errorf("Case %d: Found token when it should have been deleted.", i)
			continue
		}
	}
}
