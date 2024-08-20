package assignments

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestList(test *testing.T) {
	users := db.MustGetServerUsers()

	testCases := []struct {
		email     string
		permError bool
		course    string
	}{
		// Invalid permissions.
		{"server-user@test.edulinq.org", true, "course-without-source"},
		{"server-creator@test.edulinq.org", true, "course-without-source"},

		// Valid permissions
		// Empty
		{"server-admin@test.edulinq.org", false, "course-without-source"},
		// One Assignment
		{"server-owner@test.edulinq.org", false, "course101-with-zero-limit"},
		// Multiple Assignments
		{"course-admin@test.edulinq.org", false, "course-languages"},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"user-email": testCase.email,
			"user-pass":  util.Sha256HexFromString(*users[testCase.email].Name),
			"course-id":  testCase.course,
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint(`courses/assignments/list`), fields)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-040"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.permError {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		var responseContent ListResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		course := db.MustGetCourse(testCase.course)
		expectedInfos := core.NewAssignmentInfos(course.GetSortedAssignments())

		if !reflect.DeepEqual(expectedInfos, responseContent.Assignments) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(expectedInfos), util.MustToJSONIndent(responseContent.Assignments))
			continue
		}
	}
}
