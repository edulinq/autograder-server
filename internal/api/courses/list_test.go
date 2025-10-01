package courses

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestListBase(test *testing.T) {
	courses := db.MustGetCourses()
	expectedInfos := core.NewCourseInfosFromMap(courses)

	testCases := []struct {
		email   string
		locator string
	}{
		// Invalid Permissions
		{"server-user", "-041"},
		{"server-creator", "-041"},
		{"course-owner", "-041"},

		// Valid Permissions
		{"server-admin", ""},
		{"server-owner", ""},
	}

	for i, testCase := range testCases {
		response := core.SendTestAPIRequestFull(test, `courses/list`, nil, nil, testCase.email)
		if !response.Success {
			if testCase.locator != "" {
				if response.Locator != testCase.locator {
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

		if !reflect.DeepEqual(expectedInfos, responseContent.Courses) {
			test.Errorf("Case %d: Unexpected courses information. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(expectedInfos), util.MustToJSONIndent(responseContent.Courses))
			continue
		}
	}
}
