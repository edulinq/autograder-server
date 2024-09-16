package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestList(test *testing.T) {
	course := db.MustGetCourse(db.TEST_COURSE_ID)
	usersMap, err := db.GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Unable to get course users: '%v'.", err)
	}

	users := make([]*model.CourseUser, 0, len(usersMap))
	for _, user := range usersMap {
		users = append(users, user)
	}

	expectedInfos := core.NewCourseUserInfos(users)

	testCases := []struct {
		email     string
		permError bool
		locator   string
	}{
		// Invalid permissions.
		{"course-other", true, "-020"},
		{"course-student", true, "-020"},

		// Invalid permissions, role escalation.
		{"server-user", true, "-040"},
		{"server-creator", true, "-040"},

		// Valid permissions.
		{"course-grader", false, ""},
		{"course-admin", false, ""},
		{"course-owner", false, ""},

		// Valid permissions, role escalation.
		{"server-admin", false, ""},
		{"server-owner", false, ""},
	}

	for i, testCase := range testCases {
		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/users/list`), nil, nil, testCase.email)
		if !response.Success {
			if testCase.permError {
				if response.Locator != testCase.locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
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

		var responseContent ListResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(expectedInfos, responseContent.Users) {
			test.Errorf("Case %d: Unexpected users information. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(expectedInfos), util.MustToJSONIndent(responseContent.Users))
			continue
		}
	}
}
