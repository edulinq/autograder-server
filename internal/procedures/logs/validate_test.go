package logs

import (
	"testing"

	"github.com/edulinq/autograder/internal/log"
)

func TestValidateQueryBase(test *testing.T) {
	testCases := []struct {
		query *log.ParsedLogQuery
		// Leave empty for no error.
		expectedErrorString string
	}{
		{&log.ParsedLogQuery{}, ""},
		{&log.ParsedLogQuery{CourseID: "course101"}, ""},
		{&log.ParsedLogQuery{CourseID: "course101", AssignmentID: "hw0"}, ""},
		{&log.ParsedLogQuery{CourseID: "course101", AssignmentID: "hw0", UserEmail: "course-student@test.edulinq.org"}, ""},
		{&log.ParsedLogQuery{UserEmail: "course-student@test.edulinq.org"}, ""},

		{nil, "Cannot validate a nil query."},

		{&log.ParsedLogQuery{CourseID: "ZZZ"}, "Could not find course with ID 'ZZZ'."},
		{&log.ParsedLogQuery{CourseID: "course101", AssignmentID: "ZZZ"}, "Could not find assignment with ID 'ZZZ' in course 'course101'."},
		{&log.ParsedLogQuery{UserEmail: "ZZZ"}, "Could not find user with email 'ZZZ'."},

		{&log.ParsedLogQuery{AssignmentID: "hw0"}, "When an assignment is provided, a course must also be provided."},

		{&log.ParsedLogQuery{CourseID: "course101", AssignmentID: "hw0", UserEmail: "server-user@test.edulinq.org"}, "A course ('course101') and user ('server-user@test.edulinq.org') was provided, but the user is not enrolled in that course."},
		{&log.ParsedLogQuery{CourseID: "course101", UserEmail: "server-user@test.edulinq.org"}, "A course ('course101') and user ('server-user@test.edulinq.org') was provided, but the user is not enrolled in that course."},
	}

	for i, testCase := range testCases {
		err := validateQuery(testCase.query)

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
