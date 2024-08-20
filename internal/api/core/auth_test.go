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
		{"course-owner@test.edulinq.org", "owner", false, ""},
		{"course-admin@test.edulinq.org", "admin", false, ""},
		{"course-grader@test.edulinq.org", "grader", false, ""},
		{"course-student@test.edulinq.org", "student", false, ""},
		{"course-other@test.edulinq.org", "other", false, ""},

		{"Z", "student", false, "-013"},
		{"Zstudent@test.edulinq.org", "student", false, "-013"},
		{"course-student@test.edulinq.orgZ", "student", false, "-013"},
		{"student", "student", false, "-013"},

		{"course-student@test.edulinq.org", "", false, "-014"},
		{"course-student@test.edulinq.org", "Zstudent", false, "-014"},
		{"course-student@test.edulinq.org", "studentZ", false, "-014"},

		{"course-owner@test.edulinq.org", "owner", true, ""},
		{"course-admin@test.edulinq.org", "admin", true, ""},
		{"course-grader@test.edulinq.org", "grader", true, ""},
		{"course-student@test.edulinq.org", "student", true, ""},
		{"course-other@test.edulinq.org", "other", true, ""},

		{"Z", "student", true, "-013"},
		{"Zstudent@test.edulinq.org", "student", true, "-013"},
		{"course-student@test.edulinq.orgZ", "student", true, "-013"},
		{"student", "student", true, "-013"},

		{"course-student@test.edulinq.org", "", true, ""},
		{"course-student@test.edulinq.org", "Zstudent", true, ""},
		{"course-student@test.edulinq.org", "studentZ", true, ""},
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
