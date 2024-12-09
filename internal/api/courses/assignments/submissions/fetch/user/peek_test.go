package user

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchUserPeek(test *testing.T) {
	// There are two options, which makes for four general test cases.
	testCases := []struct {
		email            string
		targetEmail      string
		targetSubmission string
		score            float64
		foundUser        bool
		foundSubmission  bool
		permError        bool
		locator          string
	}{
		// Grader, self, recent.
		{"course-grader", "", "", -1.0, true, false, false, ""},
		{"course-grader", "course-grader@test.edulinq.org", "", -1.0, true, false, false, ""},

		// Grader, self, missing.
		{"course-grader", "", "ZZZ", -1.0, true, false, false, ""},
		{"course-grader", "course-grader@test.edulinq.org", "ZZZ", -1.0, true, false, false, ""},

		// Grader, other, recent.
		{"course-grader", "course-student@test.edulinq.org", "", 2.0, true, true, false, ""},

		// Grader, other, specific.
		{"course-grader", "course-student@test.edulinq.org", "1697406256", 0.0, true, true, false, ""},
		{"course-grader", "course-student@test.edulinq.org", "1697406265", 1.0, true, true, false, ""},
		{"course-grader", "course-student@test.edulinq.org", "1697406272", 2.0, true, true, false, ""},

		// Grader, other, specific (full ID).
		{"course-grader", "course-student@test.edulinq.org", "course101::hw0::course-student@test.edulinq.org::1697406256", 0.0, true, true, false, ""},
		{"course-grader", "course-student@test.edulinq.org", "course101::hw0::course-student@test.edulinq.org::1697406265", 1.0, true, true, false, ""},
		{"course-grader", "course-student@test.edulinq.org", "course101::hw0::course-student@test.edulinq.org::1697406272", 2.0, true, true, false, ""},

		// Grader, other, missing.
		{"course-grader", "course-student@test.edulinq.org", "ZZZ", -1.0, true, false, false, ""},

		// Grader, missing, recent.
		{"course-grader", "ZZZ@test.edulinq.org", "", -1.0, false, false, false, ""},

		// Role escalation, other, recent.
		{"server-admin", "course-student@test.edulinq.org", "", 2.0, true, true, false, ""},

		// Role escalation, other, specific.
		{"server-admin", "course-student@test.edulinq.org", "1697406256", 0.0, true, true, false, ""},

		// Role escalation, other, missing.
		{"server-admin", "course-student@test.edulinq.org", "ZZZ", -1.0, true, false, false, ""},

		// Role escalation, missing, recent.
		{"server-admin", "ZZZ@test.edulinq.org", "", -1.0, false, false, false, ""},

		// Student, self, recent.
		{"course-student", "", "", 2.0, true, true, false, ""},
		{"course-student", "course-student@test.edulinq.org", "", 2.0, true, true, false, ""},

		// Student, self, missing.
		{"course-student", "", "ZZZ", -1.0, true, false, false, ""},
		{"course-student", "course-student@test.edulinq.org", "ZZZ", -1.0, true, false, false, ""},

		// Student, other, recent.
		{"course-student", "course-grader@test.edulinq.org", "", -1.0, false, false, true, "-033"},

		// Student, other, missing.
		{"course-student", "course-grader@test.edulinq.org", "ZZZ", -1.0, false, false, true, "-033"},

		// Invalid role escalation, other, recent.
		{"server-user", "course-grader@test.edulinq.org", "", -1.0, false, false, true, "-040"},

		// Invalid role escalation, other, missing.
		{"server-user", "course-grader@test.edulinq.org", "ZZZ", -1.0, false, false, true, "-040"},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-email":      testCase.targetEmail,
			"target-submission": testCase.targetSubmission,
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/fetch/user/peek`, fields, nil, testCase.email)
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

		if !responseContent.FoundSubmission {
			continue
		}

		actualScore := responseContent.GradingInfo.Score
		if !util.IsClose(testCase.score, actualScore) {
			test.Errorf("Case %d: Unexpected submission score. Expected: '%+v', actual: '%+v'.", i, testCase.score, actualScore)
			continue
		}
	}
}
