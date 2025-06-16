package course

import (
	"maps"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchCourseScores(test *testing.T) {
	testCases := []struct {
		email       string
		targetUsers []model.CourseUserReference
		locator     string
		ids         map[string]string
	}{
		// Valid Permissions
		{"course-grader", nil, "", map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"course-admin", []model.CourseUserReference{}, "", map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"course-admin", []model.CourseUserReference{"*"}, "", map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"course-grader", []model.CourseUserReference{"student"}, "", map[string]string{
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
		}},
		{"course-grader", []model.CourseUserReference{"grader"}, "", map[string]string{
			"course-grader@test.edulinq.org": "",
		}},
		{"course-admin", []model.CourseUserReference{"*", "-grader"}, "", map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"course-admin", []model.CourseUserReference{"-*"}, "", map[string]string{}},

		// Valid Permissions, Role Escalation
		{"server-admin", []model.CourseUserReference{"*"}, "", map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"server-admin", []model.CourseUserReference{"student"}, "", map[string]string{
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
		}},
		{"server-admin", []model.CourseUserReference{"grader"}, "", map[string]string{
			"course-grader@test.edulinq.org": "",
		}},

		// Invalid Permissions
		{"course-student", []model.CourseUserReference{"*"}, "-020", nil},
		{"course-student", []model.CourseUserReference{"student"}, "-020", nil},
		{"course-other", []model.CourseUserReference{"*"}, "-020", nil},
		{"course-other", []model.CourseUserReference{"grader"}, "-020", nil},

		// Invalid Permissions, Role Escalation
		{"server-user", []model.CourseUserReference{"*"}, "-040", nil},
		{"server-user", []model.CourseUserReference{"student"}, "-040", nil},
		{"server-creator", []model.CourseUserReference{"*"}, "-040", nil},
		{"server-creator", []model.CourseUserReference{"grader"}, "-040", nil},

		// Invalid Inputs
		{"course-grader", []model.CourseUserReference{"ZZZ"}, "-636", nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-users": testCase.targetUsers,
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/fetch/course/scores`, fields, nil, testCase.email)
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

		var responseContent FetchCourseScoresResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		actualIDs := make(map[string]string, len(testCase.ids))
		for email, info := range responseContent.SubmissionInfos {
			id := ""
			if info != nil {
				id = info.ID
			}

			actualIDs[email] = id
		}

		if !maps.Equal(testCase.ids, actualIDs) {
			test.Errorf("Case %d: Submission IDs do not match. Expected: '%+v', actual: '%+v'.", i, testCase.ids, actualIDs)
			continue
		}
	}
}
