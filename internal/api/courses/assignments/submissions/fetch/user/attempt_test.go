package user

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchUserAttempt(test *testing.T) {
	// Note that computation of these paths is deferred until test time.
	studentGradingResults := map[string]*model.GradingResult{
		"1697406256": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406256")),
		"1697406265": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406265")),
		"1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
	}

	testCases := []struct {
		email            string
		targetEmail      string
		targetSubmission string
		foundUser        bool
		foundSubmission  bool
		permError        bool
		locator          string
		result           *model.GradingResult
	}{
		// Grader, self, recent.
		{"course-grader", "", "", true, false, false, "", nil},
		{"course-grader", "course-grader@test.edulinq.org", "", true, false, false, "", nil},

		// Grader, self, missing.
		{"course-grader", "", "ZZZ", true, false, false, "", nil},
		{"course-grader", "course-grader@test.edulinq.org", "ZZZ", true, false, false, "", nil},

		// Grader, other, recent.
		{"course-grader", "course-student@test.edulinq.org", "", true, true, false, "", studentGradingResults["1697406272"]},

		// Grader, other, specific.
		{"course-grader", "course-student@test.edulinq.org", "1697406256", true, true, false, "", studentGradingResults["1697406256"]},
		{"course-grader", "course-student@test.edulinq.org", "1697406265", true, true, false, "", studentGradingResults["1697406265"]},
		{"course-grader", "course-student@test.edulinq.org", "1697406272", true, true, false, "", studentGradingResults["1697406272"]},

		// Grader, other, specific (full ID).
		{"course-grader", "course-student@test.edulinq.org", "course101::hw0::course-student@test.edulinq.org::1697406256", true, true, false, "", studentGradingResults["1697406256"]},
		{"course-grader", "course-student@test.edulinq.org", "course101::hw0::course-student@test.edulinq.org::1697406265", true, true, false, "", studentGradingResults["1697406265"]},
		{"course-grader", "course-student@test.edulinq.org", "course101::hw0::course-student@test.edulinq.org::1697406272", true, true, false, "", studentGradingResults["1697406272"]},

		// Grader, other, missing.
		{"course-grader", "course-student@test.edulinq.org", "ZZZ", true, false, false, "", nil},

		// Grader, missing, recent.
		{"course-grader", "ZZZ@test.edulinq.org", "", false, false, false, "", nil},

		// Role escalation, other, recent.
		{"server-admin", "course-student@test.edulinq.org", "", true, true, false, "", studentGradingResults["1697406272"]},

		// Role escalation, other, specific.
		{"server-admin", "course-student@test.edulinq.org", "1697406256", true, true, false, "", studentGradingResults["1697406256"]},

		// Role escalation, other, missing.
		{"server-admin", "course-student@test.edulinq.org", "ZZZ", true, false, false, "", nil},

		// Role escalation, missing, recent.
		{"server-admin", "ZZZ@test.edulinq.org", "", false, false, false, "", nil},

		// Student, self, recent.
		{"course-student", "", "", true, true, false, "", studentGradingResults["1697406272"]},
		{"course-student", "course-student@test.edulinq.org", "", true, true, false, "", studentGradingResults["1697406272"]},

		// Student, self, missing.
		{"course-student", "", "ZZZ", true, false, false, "", nil},
		{"course-student", "course-student@test.edulinq.org", "ZZZ", true, false, false, "", nil},

		// Student, other, recent.
		{"course-student", "course-grader@test.edulinq.org", "", false, false, true, "-033", nil},

		// Student, other, missing.
		{"course-student", "course-grader@test.edulinq.org", "ZZZ", true, false, true, "-033", nil},

		// Invalid role escalation, other, recent.
		{"server-user", "course-grader@test.edulinq.org", "ZZZ", true, false, true, "-040", nil},

		// Invalid role escation, other, missing.
		{"server-user", "course-grader@test.edulinq.org", "ZZZ", true, false, true, "-040", nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-email":      testCase.targetEmail,
			"target-submission": testCase.targetSubmission,
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/fetch/user/attempt`, fields, nil, testCase.email)
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

		var responseContent FetchUserAttemptResponse
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

		if !reflect.DeepEqual(testCase.result, responseContent.GradingResult) {
			test.Errorf("Case %d: Unexpected submission result. Expected: '%s', actual: '%s'.", i,
				util.MustToJSONIndent(testCase.result), util.MustToJSONIndent(responseContent.GradingResult))
			continue
		}
	}
}
