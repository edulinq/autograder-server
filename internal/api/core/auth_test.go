package core

import (
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestAuth(test *testing.T) {
	type baseAPIRequest struct {
		APIRequestUserContext
		MinCourseRoleOther
	}

	testCases := []struct {
		email   string
		pass    string
		locator string
	}{
		{"course-owner@test.edulinq.org", "course-owner", ""},
		{"course-admin@test.edulinq.org", "course-admin", ""},
		{"course-grader@test.edulinq.org", "course-grader", ""},
		{"course-student@test.edulinq.org", "course-student", ""},
		{"course-other@test.edulinq.org", "course-other", ""},

		{"Z", "course-student", "-013"},
		{"Zstudent@test.edulinq.org", "course-student", "-013"},
		{"course-student@test.edulinq.orgZ", "course-student", "-013"},
		{"student", "course-student", "-013"},

		{"course-student@test.edulinq.org", "", "-014"},
		{"course-student@test.edulinq.org", "Zcourse-student", "-014"},
		{"course-student@test.edulinq.org", "course-studentZ", "-014"},

		{"root", "", "-051"},
		{"root", "root", "-051"},
	}

	for i, testCase := range testCases {
		request := baseAPIRequest{
			APIRequestUserContext: APIRequestUserContext{
				UserEmail: testCase.email,
				UserPass:  util.Sha256HexFromString(testCase.pass),
			},
		}

		apiErr := ValidateAPIRequest(nil, &request, "")

		if (apiErr == nil) && (testCase.locator != "") {
			test.Errorf("Case %d: Expecting error '%s', but got no error.", i, testCase.locator)
		} else if (apiErr != nil) && (testCase.locator == "") {
			test.Errorf("Case %d: Expecting no error, but got '%s': '%v'.", i, apiErr.Locator, apiErr)
		} else if (apiErr != nil) && (testCase.locator != "") && (apiErr.Locator != testCase.locator) {
			test.Errorf("Case %d: Got a different error than expected. Expected: '%s', actual: '%s' -- '%v'.",
				i, testCase.locator, apiErr.Locator, apiErr)
		}
	}
}
