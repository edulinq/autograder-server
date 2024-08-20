package submissions

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestRemove(test *testing.T) {
	// Leave the course in a good state after the test.
	defer db.ResetForTesting()

	testCases := []struct {
		email            string
		targetEmail      string
		targetSubmission string
		foundUser        bool
		foundSubmission  bool
		permError        bool
	}{
		// Grader, self, recent.
		{"course-grader@test.edulinq.org", "", "", true, false, false},
		{"course-grader@test.edulinq.org", "course-grader@test.edulinq.org", "", true, false, false},

		// Grader, self, missing.
		{"course-grader@test.edulinq.org", "", "ZZZ", true, false, false},
		{"course-grader@test.edulinq.org", "course-grader@test.edulinq.org", "ZZZ", true, false, false},

		// Grader, other, recent.
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", "", true, true, false},

		// Grader, other, specific.
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", "1697406256", true, true, false},
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", "1697406265", true, true, false},
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", "1697406272", true, true, false},

		// Grader, other, specific (full ID).
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406256", true, true, false},
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406265", true, true, false},
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406272", true, true, false},

		// Grader, other, missing.
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", "ZZZ", true, false, false},

		// Grader, missing, recent.
		{"course-grader@test.edulinq.org", "ZZZ@test.edulinq.org", "", false, false, false},

		// Roles below grader, other, recent.
		{"course-student@test.edulinq.org", "course-student@test.edulinq.org", "", false, false, true},
		{"course-other@test.edulinq.org", "course-student@test.edulinq.org", "", false, false, true},

		// Roles above grader, other, recent
		{"course-admin@test.edulinq.org", "course-student@test.edulinq.org", "", true, true, false},
		{"course-owner@test.edulinq.org", "course-student@test.edulinq.org", "", true, true, false},
	}

	for i, testCase := range testCases {
		// Reload the test course every time.
		db.ResetForTesting()

		fields := map[string]any{
			"target-email":      testCase.targetEmail,
			"target-submission": testCase.targetSubmission,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/assignments/submissions/remove`), fields, nil, testCase.email)

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

		var responseContent FetchUserPeekResponse
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
