package submissions

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchSubmissions(test *testing.T) {
	testCases := []struct {
		role      model.CourseUserRole
		permError bool
	}{
		{model.CourseRoleOther, true},
		{model.CourseRoleStudent, true},
		{model.CourseRoleGrader, false},
		{model.CourseRoleAdmin, false},
		{model.CourseRoleOwner, false},
	}

	submissions := map[string]*model.GradingResult{
		"other@test.com":   nil,
		"student@test.com": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
		"grader@test.com":  nil,
		"admin@test.com":   nil,
		"owner@test.com":   nil,
	}

	for i, testCase := range testCases {
		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submissions/fetch/submissions`), nil, nil, testCase.role)
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

		var responseContent FetchSubmissionsResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(submissions, responseContent.GradingResults) {
			test.Errorf("Case %d: Unexpected submission IDs. Expected: '%s', actual: '%s'.", i,
				util.MustToJSONIndent(submissions), util.MustToJSONIndent(responseContent.GradingResults))
			continue
		}
	}
}
