package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestGet(test *testing.T) {
	course := db.MustGetTestCourse()

	users, err := db.GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Failed to get course users: '%v'.", err)
	}

	testCases := []struct {
		email     string
		target    string
		permError bool
		locator   string
		expected  *model.CourseUser
	}{
		// Self.
		{"course-student", "", false, "", users["course-student@test.edulinq.org"]},
		{"course-grader", "", false, "", users["course-grader@test.edulinq.org"]},

		// Other, bad permissions.
		{"course-student", "couse-admin@test.edulinq.org", true, "-033", nil},

		// Other, good permissions.
		{"course-grader", "course-student@test.edulinq.org", false, "", users["course-student@test.edulinq.org"]},
		{"course-admin", "course-student@test.edulinq.org", false, "", users["course-student@test.edulinq.org"]},
		{"course-owner", "course-student@test.edulinq.org", false, "", users["course-student@test.edulinq.org"]},

		// Other, good permissions, role escalation.
		{"server-admin", "course-student@test.edulinq.org", false, "", users["course-student@test.edulinq.org"]},
		{"server-owner", "course-student@test.edulinq.org", false, "", users["course-student@test.edulinq.org"]},

		// Missing
		{"course-admin", "ZZZ@test.edulinq.org", false, "", nil},
		{"course-admin", "server-user@test.edulinq.org", false, "", nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-email": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/users/get`), fields, nil, testCase.email)
		if !response.Success {
			if testCase.permError {
				if testCase.locator != response.Locator {
					test.Errorf("Case %d: Incorrect error returned. Expcted '%s', found '%s'.",
						i, testCase.locator, response.Locator)
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

		var responseContent GetResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		expectedFound := (testCase.expected != nil)
		if expectedFound != responseContent.Found {
			test.Errorf("Case %d: Found user does not match. Expected: '%v', actual: '%v'.", i, expectedFound, responseContent.Found)
			continue
		}

		if testCase.expected == nil {
			continue
		}

		expectedInfo := core.NewCourseUserInfo(testCase.expected)
		if !reflect.DeepEqual(expectedInfo, responseContent.User) {
			test.Errorf("Case %d: Unexpected user result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(expectedInfo), util.MustToJSONIndent(responseContent.User))
			continue
		}
	}
}
