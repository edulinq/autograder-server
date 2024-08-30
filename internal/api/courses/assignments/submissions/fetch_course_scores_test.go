package submissions

import (
	"maps"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchCourseScores(test *testing.T) {
	testCases := []struct {
		email      string
		filterRole model.CourseUserRole
		permError  bool
		locator    string
		ids        map[string]string
	}{
		// Valid permissions
		{"course-grader", model.CourseRoleUnknown, false, "", map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"course-admin", model.CourseRoleUnknown, false, "", map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"course-grader", model.CourseRoleStudent, false, "", map[string]string{
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
		}},
		{"course-grader", model.CourseRoleGrader, false, "", map[string]string{
			"course-grader@test.edulinq.org": "",
		}},

		// Valid permissions, role escalation
		{"server-admin", model.CourseRoleUnknown, false, "", map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"server-admin", model.CourseRoleStudent, false, "", map[string]string{
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
		}},
		{"server-admin", model.CourseRoleGrader, false, "", map[string]string{
			"course-grader@test.edulinq.org": "",
		}},

		// Invalid permissions
		{"course-student", model.CourseRoleUnknown, true, "-020", nil},
		{"course-student", model.CourseRoleStudent, true, "-020", nil},
		{"course-other", model.CourseRoleUnknown, true, "-020", nil},
		{"course-other", model.CourseRoleGrader, true, "-020", nil},

		// Invalid permissions, role escalation
		{"server-user", model.CourseRoleUnknown, true, "-040", nil},
		{"server-user", model.CourseRoleStudent, true, "-040", nil},
		{"server-creator", model.CourseRoleUnknown, true, "-040", nil},
		{"server-creator", model.CourseRoleGrader, true, "-040", nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"filter-role": testCase.filterRole,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/assignments/submissions/fetch/course/scores`), fields, nil, testCase.email)
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
