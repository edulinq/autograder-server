package core

import (
	"testing"

	"github.com/edulinq/autograder/internal/config"
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
		noauth  bool
		locator string
	}{
		{"course-owner@test.edulinq.org", "course-owner", false, ""},
		{"course-admin@test.edulinq.org", "course-admin", false, ""},
		{"course-grader@test.edulinq.org", "course-grader", false, ""},
		{"course-student@test.edulinq.org", "course-student", false, ""},
		{"course-other@test.edulinq.org", "course-other", false, ""},

		{"Z", "course-student", false, "-013"},
		{"Zstudent@test.edulinq.org", "course-student", false, "-013"},
		{"course-student@test.edulinq.orgZ", "course-student", false, "-013"},
		{"student", "course-student", false, "-013"},

		{"course-student@test.edulinq.org", "", false, "-014"},
		{"course-student@test.edulinq.org", "Zcourse-student", false, "-014"},
		{"course-student@test.edulinq.org", "course-studentZ", false, "-014"},

		{"course-owner@test.edulinq.org", "course-owner", true, ""},
		{"course-admin@test.edulinq.org", "course-admin", true, ""},
		{"course-grader@test.edulinq.org", "course-grader", true, ""},
		{"course-student@test.edulinq.org", "course-student", true, ""},
		{"course-other@test.edulinq.org", "course-other", true, ""},

		{"Z", "course-student", true, "-013"},
		{"Zstudent@test.edulinq.org", "course-student", true, "-013"},
		{"course-student@test.edulinq.orgZ", "course-student", true, "-013"},
		{"student", "course-student", true, "-013"},

		{"course-student@test.edulinq.org", "", true, ""},
		{"course-student@test.edulinq.org", "Zcourse-student", true, ""},
		{"course-student@test.edulinq.org", "course-studentZ", true, ""},
	}

	oldNoAuth := config.NO_AUTH.Get()
	defer config.NO_AUTH.Set(oldNoAuth)

	for i, testCase := range testCases {
		request := baseAPIRequest{
			APIRequestUserContext: APIRequestUserContext{
				UserEmail: testCase.email,
				UserPass:  util.Sha256HexFromString(testCase.pass),
			},
		}

		config.NO_AUTH.Set(testCase.noauth)
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
