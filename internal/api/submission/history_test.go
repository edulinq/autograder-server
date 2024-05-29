package submission

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestHistory(test *testing.T) {
	testCases := []struct {
		role      model.CourseUserRole
		target    string
		found     bool
		permError bool
		expected  []*model.SubmissionHistoryItem
	}{
		// Self.
		{model.RoleStudent, "", true, false, studentHist},
		{model.RoleGrader, "", true, false, []*model.SubmissionHistoryItem{}},

		// Other
		{model.RoleGrader, "student@test.com", true, false, studentHist},
		{model.RoleStudent, "grader@test.com", true, true, nil},

		// Missing user.
		{model.RoleStudent, "ZZZ@test.com", false, true, nil},
		{model.RoleGrader, "ZZZ@test.com", false, false, []*model.SubmissionHistoryItem{}},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-email": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/history`), fields, nil, testCase.role)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-033"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
						i, expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		var responseContent HistoryResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if testCase.found != responseContent.FoundUser {
			test.Errorf("Case %d: FoundUser value mismatch. Expected: '%v', actual: '%v'.", i, testCase.found, responseContent.FoundUser)
			continue
		}

		if responseContent.History == nil {
			test.Errorf("Case %d: History is nil when is should not be: '%v'.", i, response)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, responseContent.History) {
			test.Errorf("Case %d: History does not match. Expected: '%s', actual: '%s'.", i,
				util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent.History))
			continue
		}
	}
}

var studentHist []*model.SubmissionHistoryItem = []*model.SubmissionHistoryItem{
	&model.SubmissionHistoryItem{
		ID:               "course101::hw0::student@test.com::1697406256",
		ShortID:          "1697406256",
		CourseID:         "course101",
		AssignmentID:     "hw0",
		User:             "student@test.com",
		Message:          "",
		MaxPoints:        2,
		Score:            0,
		GradingStartTime: common.MustTimestampFromString("2023-10-15T21:44:16.840060+00:00"),
	},
	&model.SubmissionHistoryItem{
		ID:               "course101::hw0::student@test.com::1697406265",
		ShortID:          "1697406265",
		CourseID:         "course101",
		AssignmentID:     "hw0",
		User:             "student@test.com",
		Message:          "",
		MaxPoints:        2,
		Score:            1,
		GradingStartTime: common.MustTimestampFromString("2023-10-15T21:44:26.445382+00:00"),
	},
	&model.SubmissionHistoryItem{
		ID:               "course101::hw0::student@test.com::1697406272",
		ShortID:          "1697406272",
		CourseID:         "course101",
		AssignmentID:     "hw0",
		User:             "student@test.com",
		Message:          "",
		MaxPoints:        2,
		Score:            2,
		GradingStartTime: common.MustTimestampFromString("2023-10-15T21:44:33.157275+00:00"),
	},
}
