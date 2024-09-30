package logs

import (
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
)

func TestCheckPermissionsBase(test *testing.T) {
	const errorMessage string = "User does not meet conditions for query."

	testCases := []struct {
		query *log.ParsedLogQuery
		email string
		// Leave empty for no error.
		expectedErrorString string
	}{
		// Sever admins can do what they want.
		{&log.ParsedLogQuery{}, "server-admin@test.edulinq.org", ""},
		{&log.ParsedLogQuery{CourseID: "course101"}, "server-admin@test.edulinq.org", ""},
		{&log.ParsedLogQuery{UserEmail: "course-student@test.edulinq.org"}, "server-admin@test.edulinq.org", ""},
		{&log.ParsedLogQuery{UserEmail: "server-owner@test.edulinq.org"}, "server-admin@test.edulinq.org", ""},
		{&log.ParsedLogQuery{UserEmail: "server-admin@test.edulinq.org"}, "server-admin@test.edulinq.org", ""},

		// Course admins can query for their course (and themselves, but not others.).
		{&log.ParsedLogQuery{}, "course-admin@test.edulinq.org", errorMessage},
		{&log.ParsedLogQuery{CourseID: "course101"}, "course-admin@test.edulinq.org", ""},
		{&log.ParsedLogQuery{UserEmail: "course-admin@test.edulinq.org"}, "course-admin@test.edulinq.org", ""},
		{&log.ParsedLogQuery{UserEmail: "course-student@test.edulinq.org"}, "course-admin@test.edulinq.org", errorMessage},

		// Course students can query for themselves only.
		{&log.ParsedLogQuery{}, "course-student@test.edulinq.org", errorMessage},
		{&log.ParsedLogQuery{CourseID: "course101"}, "course-student@test.edulinq.org", errorMessage},
		{&log.ParsedLogQuery{UserEmail: "course-student@test.edulinq.org"}, "course-student@test.edulinq.org", ""},
		{&log.ParsedLogQuery{UserEmail: "course-admin@test.edulinq.org"}, "course-student@test.edulinq.org", errorMessage},

		// Server users can query for themselves only.
		{&log.ParsedLogQuery{}, "server-user@test.edulinq.org", errorMessage},
		{&log.ParsedLogQuery{CourseID: "course101"}, "server-user@test.edulinq.org", errorMessage},
		{&log.ParsedLogQuery{UserEmail: "server-user@test.edulinq.org"}, "server-user@test.edulinq.org", ""},
		{&log.ParsedLogQuery{UserEmail: "course-student@test.edulinq.org"}, "server-user@test.edulinq.org", errorMessage},

		// Nil checks.
		{nil, "server-admin@test.edulinq.org", "Cannot check log permissions with a nil query."},
		{&log.ParsedLogQuery{}, "ZZZ@test.edulinq.org", "Cannot check log permissions with a nil user."},
	}

	for i, testCase := range testCases {
		user := db.MustGetServerUser(testCase.email)

		err := checkPermissions(testCase.query, user)

		actualErrorString := ""
		if err != nil {
			if testCase.expectedErrorString == "" {
				test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
				continue
			}

			actualErrorString = err.Error()
		}

		if testCase.expectedErrorString != actualErrorString {
			test.Errorf("Case %d: Incorrect error string. Expected: '%s', Actual: '%s'.", i, testCase.expectedErrorString, actualErrorString)
			continue
		}
	}
}
