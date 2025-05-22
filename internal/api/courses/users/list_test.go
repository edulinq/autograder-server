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
	db.ResetForTesting()
	defer db.ResetForTesting()

	course := db.MustGetTestCourse()

	usersMap, err := db.GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Unable to get course users: '%v'.", err)
	}

	allUsers := make([]*model.CourseUser, 0, len(usersMap))
	for _, user := range usersMap {
		allUsers = append(allUsers, user)
	}

	testCases := []struct {
		email          string
		input          []string
		expectedUsers  []*model.CourseUser
		expectedErrors map[string]string
		locator        string
	}{
		// Invalid Permissions
		{"course-other", nil, nil, nil, "-020"},
		{"course-student", nil, nil, nil, "-020"},

		// Invalid Permissions, Role Escalation
		{"server-user", nil, nil, nil, "-040"},
		{"server-creator", nil, nil, nil, "-040"},

		// Valid Permissions, All Users
		{"course-grader", nil, allUsers, map[string]string{}, ""},
		{"course-admin", []string{}, allUsers, map[string]string{}, ""},
		{"course-owner", []string{"*"}, allUsers, map[string]string{}, ""},

		// Valid Permissions, Role Escalation, All Users
		{
			"server-admin",
			[]string{"admin", "grader", "other", "owner", "student"},
			allUsers,
			map[string]string{},
			"",
		},
		{
			"server-owner",
			[]string{
				"course-admin@test.edulinq.org",
				"course-grader@test.edulinq.org",
				"course-other@test.edulinq.org",
				"course-owner@test.edulinq.org",
				"course-student@test.edulinq.org",
			},
			allUsers,
			map[string]string{},
			"",
		},

		// No Users
		{"course-admin", []string{"-*"}, []*model.CourseUser{}, map[string]string{}, ""},

		// Invalid Users
		{
			"course-admin",
			[]string{"server-admin@test.edulinq.org"},
			nil,
			map[string]string{
				"server-admin@test.edulinq.org": "Unknown course user: 'server-admin@test.edulinq.org'.",
			},
			"",
		},
		{
			"course-admin",
			[]string{"creator"},
			nil,
			map[string]string{
				"creator": "Unknown course role: 'creator'.",
			},
			"",
		},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"users": testCase.input,
		}

		response := core.SendTestAPIRequestFull(test, `courses/users/list`, fields, nil, testCase.email)
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

		if !reflect.DeepEqual(testCase.expectedErrors, responseContent.UserErrors) {
			test.Errorf("Case %d: Unexpected user errors. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expectedErrors), util.MustToJSONIndent(responseContent.UserErrors))
			continue
		}

		expectedInfos := core.NewCourseUserInfos(testCase.expectedUsers)

		if !reflect.DeepEqual(expectedInfos, responseContent.Users) {
			test.Errorf("Case %d: Unexpected users information. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(expectedInfos), util.MustToJSONIndent(responseContent.Users))
			continue
		}
	}
}

func TestListEmptyCourse(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	course := db.MustGetTestCourse()

	users, err := db.GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Error when getting course users: '%v'.", err)
	}

	for _, user := range users {
		_, _, err = db.RemoveUserFromCourse(course, user.Email)
		if err != nil {
			test.Fatalf("Error when removing the user: '%v'.", err)
		}
	}

	expectedInfos := core.NewCourseUserInfos([]*model.CourseUser{})

	response := core.SendTestAPIRequestFull(test, `courses/users/list`, nil, nil, "server-admin")
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent ListResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if !reflect.DeepEqual(expectedInfos, responseContent.Users) {
		test.Fatalf("Unexpected users information. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expectedInfos), util.MustToJSONIndent(responseContent.Users))
	}
}
