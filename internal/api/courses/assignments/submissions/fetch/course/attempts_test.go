package course

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchCourseAttempts(test *testing.T) {
	submissions := map[string]*model.GradingResult{
		"course-other@test.edulinq.org":   nil,
		"course-student@test.edulinq.org": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
		"course-grader@test.edulinq.org":  nil,
		"course-admin@test.edulinq.org":   nil,
		"course-owner@test.edulinq.org":   nil,
	}

	testCases := []struct {
		email       string
		targetUsers []model.CourseUserReference
		locator     string
		output      map[string]*model.GradingResult
	}{
		// Invalid Permissions
		{"course-other", nil, "-020", nil},
		{"course-student", nil, "-020", nil},

		// Invalid Permissions, Role Escalation
		{"server-user", nil, "-040", nil},
		{"server-creator", nil, "-040", nil},

		// Valid Permissions
		{"course-grader", nil, "", submissions},
		{"course-admin", []model.CourseUserReference{}, "", submissions},
		{"course-owner", []model.CourseUserReference{"*"}, "", submissions},

		{"course-owner", []model.CourseUserReference{"student"}, "", map[string]*model.GradingResult{
			"course-student@test.edulinq.org": submissions["course-student@test.edulinq.org"],
		}},

		// Valid Permissions, Role Escalation
		{"server-admin", []model.CourseUserReference{"*"}, "", submissions},
		{"server-owner", []model.CourseUserReference{"*"}, "", submissions},

		// Invalid Inputs
		{"course-grader", []model.CourseUserReference{"ZZZ"}, "-637", nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-users": testCase.targetUsers,
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/fetch/course/attempts`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.locator != "" {
				if response.Locator != testCase.locator {
					test.Errorf("Case %d: Incorrect error returned on permissions error. Expected '%s', found '%s'.",
						i, testCase.locator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.locator != "" {
			test.Errorf("Case %d: Did not get an expected error: '%s'.", i, testCase.locator)
			continue
		}

		var responseContent FetchCourseAttemptsResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(testCase.output, responseContent.GradingResults) {
			test.Errorf("Case %d: Unexpected submission IDs. Expected: '%s', actual: '%s'.", i,
				util.MustToJSONIndent(testCase.output), util.MustToJSONIndent(responseContent.GradingResults))
			continue
		}
	}
}
