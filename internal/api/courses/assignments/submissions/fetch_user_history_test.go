package submissions

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestHistory(test *testing.T) {
	testCases := []struct {
		email     string
		target    string
		found     bool
		permError bool
		locator   string
		expected  []*model.SubmissionHistoryItem
	}{
		// Self.
		{"course-student", "", true, false, "", studentHist},
		{"course-grader", "", true, false, "", []*model.SubmissionHistoryItem{}},

		// Other
		{"course-grader", "course-student@test.edulinq.org", true, false, "", studentHist},
		{"course-student", "course-grader@test.edulinq.org", true, true, "-033", nil},

		// Other, role escalation
		{"server-admin", "course-student@test.edulinq.org", true, false, "", studentHist},
		{"server-owner", "course-student@test.edulinq.org", true, false, "", studentHist},

		// Invalid role escalation
		{"server-user", "", false, true, "-040", nil},
		{"server-creator", "", false, true, "-040", nil},

		// Missing user.
		{"course-student", "ZZZ@test.edulinq.org", false, true, "-033", nil},
		{"course-grader", "ZZZ@test.edulinq.org", false, false, "", []*model.SubmissionHistoryItem{}},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-email": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/fetch/user/history`, fields, nil, testCase.email)
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

		var responseContent FetchUserHistoryResponse
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
		ID:               "course101::hw0::course-student@test.edulinq.org::1697406256",
		ShortID:          "1697406256",
		CourseID:         "course101",
		AssignmentID:     "hw0",
		User:             "course-student@test.edulinq.org",
		Message:          "",
		MaxPoints:        2,
		Score:            0,
		GradingStartTime: timestamp.Timestamp(1697406256000),
	},
	&model.SubmissionHistoryItem{
		ID:               "course101::hw0::course-student@test.edulinq.org::1697406265",
		ShortID:          "1697406265",
		CourseID:         "course101",
		AssignmentID:     "hw0",
		User:             "course-student@test.edulinq.org",
		Message:          "",
		MaxPoints:        2,
		Score:            1,
		GradingStartTime: timestamp.Timestamp(1697406266000),
	},
	&model.SubmissionHistoryItem{
		ID:               "course101::hw0::course-student@test.edulinq.org::1697406272",
		ShortID:          "1697406272",
		CourseID:         "course101",
		AssignmentID:     "hw0",
		User:             "course-student@test.edulinq.org",
		Message:          "",
		MaxPoints:        2,
		Score:            2,
		GradingStartTime: timestamp.Timestamp(1697406273000),
	},
}
