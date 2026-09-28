package status

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestList(test *testing.T) {
	testCases := []struct {
		email   string
		target  string
		locator string
	}{
		// Base
		{"course-other", "course101", ""},
		{"course-student", "course101", ""},
		{"course-grader", "course101", ""},
		{"course-admin", "course101", ""},
		{"course-owner", "course101", ""},
		{"server-admin", "course101", ""},
		{"server-owner", "course101", ""},

		// Bad Perms
		{"server-user", "course101", "-040"},
		{"server-creator", "course101", "-040"},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"course-id": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, `courses/status/list`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.locator != "" {
				if testCase.locator != response.Locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
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

		var responseContent ListResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		expectedStatuses := db.MustGetCourse(testCase.target).Statuses

		if !reflect.DeepEqual(expectedStatuses, responseContent.Statuses) {
			test.Fatalf("Case %d: Unexpected statuses. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(expectedStatuses), util.MustToJSONIndent(responseContent.Statuses))
			continue
		}
	}

}
