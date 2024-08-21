package submissions

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchUserAttempts(test *testing.T) {
	// Note that computation of these paths is deferred until test time.
	studentGradingResults := []*model.GradingResult{
		model.MustLoadGradingResult(getTestSubmissionResultPath("1697406256")),
		model.MustLoadGradingResult(getTestSubmissionResultPath("1697406265")),
		model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
	}

	testCases := []struct {
		email       string
		targetEmail string
		foundUser   bool
		permError   bool
		locator     string
		result      []*model.GradingResult
	}{
		// Grader, self.
		{"course-grader@test.edulinq.org", "", true, false, "", []*model.GradingResult{}},
		{"course-grader@test.edulinq.org", "course-grader@test.edulinq.org", true, false, "", []*model.GradingResult{}},

		// Grader, other.
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", true, false, "", studentGradingResults},

		// Grader, missing.
		{"course-grader@test.edulinq.org", "ZZZ@test.edulinq.org", false, false, "", []*model.GradingResult{}},

		// Role escalation, other.
		{"server-admin@test.edulinq.org", "course-student@test.edulinq.org", true, false, "", studentGradingResults},

		// Invalid role escalation.
		{"server-user@test.edulinq.org", "", true, true, "-040", nil},

		// Student, self.
		{"course-student@test.edulinq.org", "", true, true, "-020", nil},
	}

	for i, testCase := range testCases {
		field := map[string]any{
			"target-email": testCase.targetEmail,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/assignments/submissions/fetch/user/attempts`), field, nil, testCase.email)
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

		var responseContent FetchUserAttemptsResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if testCase.foundUser != responseContent.FoundUser {
			test.Errorf("Case %d: Found user does not match. Expected: '%v', actual: '%v'.", i, testCase.foundUser, responseContent.FoundUser)
			continue
		}

		if !reflect.DeepEqual(testCase.result, responseContent.GradingResults) {
			test.Errorf("Case %d: Unexpected submission result. Expected: '%s', actual: '%s'.", i,
				util.MustToJSONIndent(testCase.result), util.MustToJSONIndent(responseContent.GradingResults))
			continue
		}
	}
}
