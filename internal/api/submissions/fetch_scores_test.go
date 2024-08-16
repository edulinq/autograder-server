package submissions

import (
	"maps"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetchScores(test *testing.T) {
	testCases := []struct {
		role       model.CourseUserRole
		filterRole model.CourseUserRole
		permError  bool
		ids        map[string]string
	}{
		{model.CourseRoleGrader, model.CourseRoleUnknown, false, map[string]string{
			"other@test.com":   "",
			"student@test.com": "course101::hw0::student@test.com::1697406272",
			"grader@test.com":  "",
			"admin@test.com":   "",
			"owner@test.com":   "",
			"root":             "",
		}},
		{model.CourseRoleAdmin, model.CourseRoleUnknown, false, map[string]string{
			"other@test.com":   "",
			"student@test.com": "course101::hw0::student@test.com::1697406272",
			"grader@test.com":  "",
			"admin@test.com":   "",
			"owner@test.com":   "",
			"root":             "",
		}},
		{model.CourseRoleGrader, model.CourseRoleStudent, false, map[string]string{
			"student@test.com": "course101::hw0::student@test.com::1697406272",
			"root":             "",
		}},
		{model.CourseRoleGrader, model.CourseRoleGrader, false, map[string]string{
			"grader@test.com": "",
			"root":            "",
		}},
		{model.CourseRoleStudent, model.CourseRoleUnknown, true, nil},
		{model.CourseRoleStudent, model.CourseRoleStudent, true, nil},
		{model.CourseRoleOther, model.CourseRoleUnknown, true, nil},
		{model.CourseRoleOther, model.CourseRoleGrader, true, nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"filter-role": testCase.filterRole,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submissions/fetch/scores`), fields, nil, testCase.role)
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

		var responseContent FetchScoresResponse
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
