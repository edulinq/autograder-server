package submissions

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchCourseAttemps(test *testing.T) {
	testCases := []struct {
		email     string
		permError bool
	}{
		{"course-other@test.edulinq.org", true},
		{"course-student@test.edulinq.org", true},
		{"course-grader@test.edulinq.org", false},
		{"course-admin@test.edulinq.org", false},
		{"course-owner@test.edulinq.org", false},
	}

	submissions := map[string]*model.GradingResult{
		"course-other@test.edulinq.org":   nil,
		"course-student@test.edulinq.org": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
		"course-grader@test.edulinq.org":  nil,
		"course-admin@test.edulinq.org":   nil,
		"course-owner@test.edulinq.org":   nil,
	}

	for i, testCase := range testCases {
		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/assignments/submissions/fetch/course/attemps`), nil, nil, testCase.email)
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

		var responseContent FetchCourseAttempsResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(submissions, responseContent.GradingResults) {
			test.Errorf("Case %d: Unexpected submission IDs. Expected: '%s', actual: '%s'.", i,
				util.MustToJSONIndent(submissions), util.MustToJSONIndent(responseContent.GradingResults))
			continue
		}
	}
}
