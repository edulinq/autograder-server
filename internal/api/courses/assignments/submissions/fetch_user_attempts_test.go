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
		role        model.CourseUserRole
		targetEmail string
		foundUser   bool
		permError   bool
		result      []*model.GradingResult
	}{
		// Grader, self.
		{model.CourseRoleGrader, "", true, false, []*model.GradingResult{}},
		{model.CourseRoleGrader, "grader@test.com", true, false, []*model.GradingResult{}},

		// Grader, other.
		{model.CourseRoleGrader, "student@test.com", true, false, studentGradingResults},

		// Grader, missing.
		{model.CourseRoleGrader, "ZZZ@test.com", false, false, []*model.GradingResult{}},

		// Student, self.
		{model.CourseRoleStudent, "", true, true, nil},
	}

	for i, testCase := range testCases {
		field := map[string]any{
			"target-email": testCase.targetEmail,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/assignments/submissions/fetch/user/attempts`), field, nil, testCase.role)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-020"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
						i, expectedLocator, response.Locator)
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
