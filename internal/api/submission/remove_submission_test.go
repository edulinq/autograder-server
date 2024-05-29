package submission

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestRemoveSubmission(test *testing.T) {
	// Leave the course in a good state after the test.
	defer db.ResetForTesting()

	testCases := []struct {
		role             model.CourseUserRole
		targetEmail      string
		targetSubmission string
		foundUser        bool
		foundSubmission  bool
		permError        bool
	}{
		// Grader, self, recent.
		{model.RoleGrader, "", "", true, false, false},
		{model.RoleGrader, "grader@test.com", "", true, false, false},

		// Grader, self, missing.
		{model.RoleGrader, "", "ZZZ", true, false, false},
		{model.RoleGrader, "grader@test.com", "ZZZ", true, false, false},

		// Grader, other, recent.
		{model.RoleGrader, "student@test.com", "", true, true, false},

		// Grader, other, specific.
		{model.RoleGrader, "student@test.com", "1697406256", true, true, false},
		{model.RoleGrader, "student@test.com", "1697406265", true, true, false},
		{model.RoleGrader, "student@test.com", "1697406272", true, true, false},

		// Grader, other, specific (full ID).
		{model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406256", true, true, false},
		{model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406265", true, true, false},
		{model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406272", true, true, false},

		// Grader, other, missing.
		{model.RoleGrader, "student@test.com", "ZZZ", true, false, false},

		// Grader, missing, recent.
		{model.RoleGrader, "ZZZ@test.com", "", false, false, false},

		// Roles below grader, other, recent.
		{model.RoleStudent, "student@test.com", "", false, false, true},
		{model.RoleOther, "student@test.com", "", false, false, true},

		// Roles above grader, other, recent
		{model.RoleAdmin, "student@test.com", "", true, true, false},
		{model.RoleOwner, "student@test.com", "", true, true, false},
	}

	for i, testCase := range testCases {
		// Reload the test course every time.
		db.ResetForTesting()

		fields := map[string]any{
			"target-email":      testCase.targetEmail,
			"target-submission": testCase.targetSubmission,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/remove`), fields, nil, testCase.role)

		if !response.Success {
			if testCase.permError {
				expectedLocator := "-020"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned on permissions error. Expected '%s', found '%s'.",
						i, expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		var responseContent PeekResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if testCase.foundUser != responseContent.FoundUser {
			test.Errorf("Case %d: Found user does not match. Expected: '%v', actual: '%v'.", i, testCase.foundUser, responseContent.FoundUser)
			continue
		}

		if testCase.foundSubmission != responseContent.FoundSubmission {
			test.Errorf("Case %d: Found submission does not match. Expected: '%v', actual: '%v'.", i, testCase.foundSubmission, responseContent.FoundSubmission)
			continue
		}
	}
}
