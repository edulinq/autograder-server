package users

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
	"reflect"
	"testing"
)

func TestUserDrop(test *testing.T) {
	defer db.ResetForTesting()

	course := db.MustGetTestCourse()

	testCases := []struct {
		email     string
		target    string
		permError bool
		locator   string
		expected  DropResponse
	}{
		// Valid permissions.
		{"course-admin", "course-other@test.edulinq.org", false, "", DropResponse{true}},
		{"course-admin", "course-student@test.edulinq.org", false, "", DropResponse{true}},
		{"course-owner", "course-grader@test.edulinq.org", false, "", DropResponse{true}},
		{"course-owner", "course-admin@test.edulinq.org", false, "", DropResponse{true}},

		// Valid permissions, role escalation.
		{"server-admin", "course-other@test.edulinq.org", false, "", DropResponse{true}},
		{"server-admin", "course-student@test.edulinq.org", false, "", DropResponse{true}},
		{"server-admin", "course-grader@test.edulinq.org", false, "", DropResponse{true}},
		{"server-admin", "course-admin@test.edulinq.org", false, "", DropResponse{true}},
		{"server-admin", "course-owner@test.edulinq.org", false, "", DropResponse{true}},

		// Invalid permissions.
		{"course-student", "course-other@test.edulinq.org", true, "-020", DropResponse{false}},
		{"course-grader", "course-other@test.edulinq.org", true, "-020", DropResponse{false}},

		// Invalid permissions, role escalation.
		{"server-creator", "course-other@test.edulinq.org", true, "-040", DropResponse{false}},
		{"server-user", "course-other@test.edulinq.org", true, "-040", DropResponse{false}},

		// Target user is not found.
		{"course-owner", "ZZZ", false, "", DropResponse{false}},

		// Target user not enrolled in course.
		{"course-owner", "server-user@test.edulinq.org", false, "", DropResponse{false}},

		// Complex invalid permissions, removing someone with equal or higher course roles.
		{"course-admin", "course-admin@test.edulinq.org", true, "-613", DropResponse{false}},
		{"course-admin", "course-owner@test.edulinq.org", true, "-613", DropResponse{false}},
		{"course-owner", "course-owner@test.edulinq.org", true, "-613", DropResponse{false}},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		initialUser, err := db.GetCourseUser(course, testCase.target)
		if err != nil {
			test.Errorf("Case %d: Failed to get initial user: '%v'.", i, err)
			continue
		}

		fields := map[string]any{
			"target-email": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, core.makeFullAPIPath(`courses/users/drop`), fields, nil, testCase.email)

		if !response.Success {
			if testCase.permError {
				if testCase.locator != response.Locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.locator, response.Locator)
					continue
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
				continue
			}
		}

		var responseContent DropResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if testCase.expected != responseContent {
			test.Errorf("Case %d: Unexpected drop result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}

		user, err := db.GetCourseUser(course, testCase.target)
		if err != nil {
			test.Errorf("Case %d: Failed to get user: '%v'.", i, err)
			continue
		}

		if !testCase.expected.FoundUser {
			// Ensure that the user is NOT removed.
			if !reflect.DeepEqual(initialUser, user) {
				test.Errorf("Case %d: Unexpected user change. Expected: '%+v', actual: '%+v'.",
					i, initialUser, user)
				continue
			}
		} else {
			// Ensure that the user is removed.
			if user != nil {
				test.Errorf("Case %d: User not dropped from course: '%+v'.", i, user)
				continue
			}
		}
	}
}
