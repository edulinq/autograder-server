package submissions

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchCourseAttempts(test *testing.T) {
	testCases := []struct {
		email     string
		permError bool
		locator   string
	}{
		// Invalid permissions
		{"course-other", true, "-020"},
		{"course-student", true, "-020"},

		// Invalid permissions, role escalation
		{"server-user", true, "-040"},
		{"server-creator", true, "-040"},

		// Valid permissions
		{"course-grader", false, ""},
		{"course-admin", false, ""},
		{"course-owner", false, ""},

		// Valid permissions, role escalation
		{"server-admin", false, ""},
		{"server-owner", false, ""},
	}

	submissions := map[string]*model.GradingResult{
		"course-other@test.edulinq.org":   nil,
		"course-student@test.edulinq.org": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
		"course-grader@test.edulinq.org":  nil,
		"course-admin@test.edulinq.org":   nil,
		"course-owner@test.edulinq.org":   nil,
	}

	for i, testCase := range testCases {
		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/fetch/course/attempts`, nil, nil, testCase.email)
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

		var responseContent FetchCourseAttemptsResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(submissions, responseContent.GradingResults) {
			test.Errorf("Case %d: Unexpected submission IDs. Expected: '%s', actual: '%s'.", i,
				util.MustToJSONIndent(submissions), util.MustToJSONIndent(responseContent.GradingResults))
			continue
		}
	}
}
