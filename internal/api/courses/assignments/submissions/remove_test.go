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
		locator          string
	}{
		// Grader, self, recent.
		{"course-grader", "", "", true, false, false, ""},
		{"course-grader", "course-grader@test.edulinq.org", "", true, false, false, ""},

		// Grader, self, missing.
		{"course-grader", "", "ZZZ", true, false, false, ""},
		{"course-grader", "course-grader@test.edulinq.org", "ZZZ", true, false, false, ""},

		// Grader, other, recent.
		{"course-grader", "course-student@test.edulinq.org", "", true, true, false, ""},

		// Grader, other, specific.
		{"course-grader", "course-student@test.edulinq.org", "1697406256", true, true, false, ""},
		{"course-grader", "course-student@test.edulinq.org", "1697406265", true, true, false, ""},
		{"course-grader", "course-student@test.edulinq.org", "1697406272", true, true, false, ""},

		// Grader, other, specific (full ID).
		{"course-grader", "course-student@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406256", true, true, false, ""},
		{"course-grader", "course-student@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406265", true, true, false, ""},
		{"course-grader", "course-student@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406272", true, true, false, ""},

		// Grader, other, missing.
		{"course-grader", "course-student@test.edulinq.org", "ZZZ", true, false, false, ""},

		// Grader, missing, recent.
		{"course-grader", "ZZZ@test.edulinq.org", "", false, false, false, ""},

		// Roles below grader, other, recent.
		{"course-student", "course-student@test.edulinq.org", "", false, false, true, "-020"},
		{"course-other", "course-student@test.edulinq.org", "", false, false, true, "-020"},

		// Roles above grader, other, recent
		{"course-admin", "course-student@test.edulinq.org", "", true, true, false, ""},
		{"course-owner", "course-student@test.edulinq.org", "", true, true, false, ""},

		// Role escalation, other, recent
		{"server-admin", "course-student@test.edulinq.org", "", true, true, false, ""},
		{"server-owner", "course-student@test.edulinq.org", "", true, true, false, ""},

		// Invalid role escalation, other, recent
		{"server-user", "course-student@test.edulinq.org", "", false, false, true, "-040"},
		{"server-creator", "course-student@test.edulinq.org", "", false, false, true, "-040"},
	}

	for i, testCase := range testCases {
		// Reload the test course every time.
		db.ResetForTesting()

		fields := map[string]any{
			"target-email":      testCase.targetEmail,
			"target-submission": testCase.targetSubmission,
		}

		response := core.SendTestAPIRequestFull(test, core.makeFullAPIPath(`courses/assignments/submissions/remove`), fields, nil, testCase.email)

		if !response.Success {
			if testCase.permError {
				if response.Locator != testCase.locator {
					test.Errorf("Case %d: Incorrect error returned on permissions error. Expected '%s', found '%s'.",
						i, testCase.locator, response.Locator)
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
