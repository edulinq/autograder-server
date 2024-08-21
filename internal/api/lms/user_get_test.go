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
		email     string
		target    string
		permError bool
		expected  *core.CourseUserInfo
	}{
		{"course-grader@test.edulinq.org", "course-other@test.edulinq.org", false, &core.CourseUserInfo{"course-other@test.edulinq.org", "course-other", model.CourseRoleOther, "lms-course-other@test.edulinq.org"}},
		{"course-grader@test.edulinq.org", "course-student@test.edulinq.org", false, &core.CourseUserInfo{"course-student@test.edulinq.org", "course-student", model.CourseRoleStudent, "lms-course-student@test.edulinq.org"}},
		{"course-grader@test.edulinq.org", "course-grader@test.edulinq.org", false, &core.CourseUserInfo{"course-grader@test.edulinq.org", "course-grader", model.CourseRoleGrader, "lms-course-grader@test.edulinq.org"}},
		{"course-grader@test.edulinq.org", "course-admin@test.edulinq.org", false, &core.CourseUserInfo{"course-admin@test.edulinq.org", "course-admin", model.CourseRoleAdmin, "lms-course-admin@test.edulinq.org"}},
		{"course-grader@test.edulinq.org", "course-owner@test.edulinq.org", false, &core.CourseUserInfo{"course-owner@test.edulinq.org", "course-owner", model.CourseRoleOwner, "lms-course-owner@test.edulinq.org"}},

		{"course-student@test.edulinq.org", "course-student@test.edulinq.org", true, nil},

		{"course-grader@test.edulinq.org", "", false, nil},

		{"course-grader@test.edulinq.org", "ZZZ", false, nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-email": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`lms/user/get`), fields, nil, testCase.email)
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
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
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
