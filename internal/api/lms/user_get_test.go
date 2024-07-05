package lms

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestUserGet(test *testing.T) {
	testCases := []struct {
		role      model.CourseUserRole
		target    string
		permError bool
		expected  *core.UserInfo
	}{
		{model.CourseRoleGrader, "other@test.com", false, &core.UserInfo{"other@test.com", "other", model.CourseRoleOther, "lms-other@test.com"}},
		{model.CourseRoleGrader, "student@test.com", false, &core.UserInfo{"student@test.com", "student", model.CourseRoleStudent, "lms-student@test.com"}},
		{model.CourseRoleGrader, "grader@test.com", false, &core.UserInfo{"grader@test.com", "grader", model.CourseRoleGrader, "lms-grader@test.com"}},
		{model.CourseRoleGrader, "admin@test.com", false, &core.UserInfo{"admin@test.com", "admin", model.CourseRoleAdmin, "lms-admin@test.com"}},
		{model.CourseRoleGrader, "owner@test.com", false, &core.UserInfo{"owner@test.com", "owner", model.CourseRoleOwner, "lms-owner@test.com"}},

		{model.CourseRoleStudent, "student@test.com", true, nil},

		{model.CourseRoleGrader, "", false, nil},

		{model.CourseRoleGrader, "ZZZ", false, nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-email": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`lms/user/get`), fields, nil, testCase.role)
		if !response.Success {
			expectedLocator := ""
			if testCase.permError {
				expectedLocator = "-020"
			} else if testCase.target == "" {
				expectedLocator = "-034"
			}

			if expectedLocator == "" {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			} else {
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expcted '%s', found '%s'.",
						i, expectedLocator, response.Locator)
				}
			}

			continue
		}

		var responseContent UserGetResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		expectedFound := (testCase.expected != nil)
		if expectedFound != responseContent.FoundAGUser {
			test.Errorf("Case %d: Found AG user does not match. Expected: '%v', actual: '%v'.", i, expectedFound, responseContent.FoundAGUser)
			continue
		}

		if expectedFound != responseContent.FoundLMSUser {
			test.Errorf("Case %d: Found LMS user does not match. Expected: '%v', actual: '%v'.", i, expectedFound, responseContent.FoundLMSUser)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, responseContent.User) {
			test.Errorf("Case %d: Unexpected user result. Expected: '%+v', actual: '%+v'.", i, testCase.expected, responseContent.User)
			continue
		}
	}
}
