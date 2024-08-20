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
		ids        map[string]string
	}{
		{"course-grader@test.edulinq.org", model.CourseRoleUnknown, false, map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"course-admin@test.edulinq.org", model.CourseRoleUnknown, false, map[string]string{
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		}},
		{"course-grader@test.edulinq.org", model.CourseRoleStudent, false, map[string]string{
			"course-student@test.edulinq.org": "course101::hw0::student@test.edulinq.org::1697406272",
		}},
		{"course-grader@test.edulinq.org", model.CourseRoleGrader, false, map[string]string{
			"course-grader@test.edulinq.org": "",
		}},
		{"course-student@test.edulinq.org", model.CourseRoleUnknown, true, nil},
		{"course-student@test.edulinq.org", model.CourseRoleStudent, true, nil},
		{"course-other@test.edulinq.org", model.CourseRoleUnknown, true, nil},
		{"course-other@test.edulinq.org", model.CourseRoleGrader, true, nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"filter-role": testCase.filterRole,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/assignments/submissions/fetch/course/scores`), fields, nil, testCase.email)
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
