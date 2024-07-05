package db

import (
	"reflect"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

func (this *DBTests) DBTestResolveCourseUsers(test *testing.T) {
	defer ResetForTesting()

	oldValue := log.SetBackgroundLogging(false)
	defer log.SetBackgroundLogging(oldValue)

	log.SetLevels(log.LevelOff, log.LevelWarn)
	defer log.SetLevelFatal()

	// Wait for old logs to get written.
	time.Sleep(10 * time.Millisecond)

	Clear()
	defer Clear()

	testCases := []struct {
		input          []string
		expectedOutput []string
		addUsers       []*model.CourseUser
		removeUsers    []string
		numWarnings    int
	}{
		// This test case tests the empty slice input.
		{
			[]string{},
			[]string{},
			nil,
			[]string{},
			0,
		},

		// This is a simple test case for the empty string input.
		{
			[]string{""},
			[]string{},
			nil,
			[]string{},
			0,
		},

		// This is a test to ensure the output is sorted.
		{
			[]string{"b@test.com", "a@test.com", "c@test.com"},
			[]string{"a@test.com", "b@test.com", "c@test.com"},
			nil,
			[]string{},
			0,
		},

		// This is a test to ensure miscapitalized emails only get returned once.
		{
			[]string{"a@test.com", "A@tesT.CoM", "A@TESt.COM"},
			[]string{"a@test.com"},
			nil,
			[]string{},
			0,
		},

		// This is a basic test to ensure that a role gets mapped to the correct email.
		{
			[]string{"admin"},
			[]string{"admin@test.com"},
			nil,
			[]string{},
			0,
		},

		// This is a test for our all roles character, the *.
		{
			[]string{"*"},
			[]string{"admin@test.com", "grader@test.com", "other@test.com", "owner@test.com", "student@test.com"},
			nil,
			[]string{},
			0,
		},

		// This test case is given redundant roles and emails.
		// It tests to ensures we do not produce duplicates on this input.
		{
			[]string{"other", "*", "grader@test.com"},
			[]string{"admin@test.com", "grader@test.com", "other@test.com", "owner@test.com", "student@test.com"},
			nil,
			[]string{},
			0,
		},

		// This test case tests if miscapitalized roles still function.
		{
			[]string{"OTHER"},
			[]string{"other@test.com"},
			nil,
			[]string{},
			0,
		},

		// This test case tests if warnings are issued on invalid roles.
		{
			[]string{"trash", "garbage", "waste", "recycle!"},
			[]string{},
			nil,
			[]string{},
			4,
		},

		// This test adds new Users to the course and ensures we retrieve all emails for the given role.
		{
			[]string{"student"},
			[]string{"a_student@test.com", "b_student@test.com", "student@test.com"},
			[]*model.CourseUser{
				&model.CourseUser{"a_student@test.com", nil, model.CourseRoleStudent, nil},
				&model.CourseUser{"b_student@test.com", nil, model.CourseRoleStudent, nil},
			},
			[]string{},
			0,
		},

		// This is a test case to see if we properly trim whitespace.
		{
			[]string{"\t\n student    ", "\n \t testing@test.com", "\t\n     \t    \n"},
			[]string{"student@test.com", "testing@test.com"},
			nil,
			[]string{},
			0,
		},

		// This test case removes the only user from the "owner" role, so we check that a role without any users still functions properly.
		{
			[]string{"owner", "student"},
			[]string{"student@test.com"},
			nil,
			[]string{"owner@test.com"},
			0,
		},

		// This test supplies a single role that resolves to nothing.
		{
			[]string{"owner"},
			[]string{},
			nil,
			[]string{"owner@test.com"},
			0,
		},
	}

	for i, testCase := range testCases {
		ResetForTesting()
		course := MustGetCourse(TEST_COURSE_ID)

		for _, newUser := range testCase.addUsers {
			UpsertCourseUser(course, newUser)
		}

		for _, removeUser := range testCase.removeUsers {
			RemoveUserFromCourse(course, removeUser)
		}

		actualOutput, err := ResolveCourseUsers(course, testCase.input)
		if err != nil {
			test.Errorf("Case %d (%+v): Resolve user failed: '%v'.", i, testCase, err)
			continue
		}

		if !reflect.DeepEqual(testCase.expectedOutput, actualOutput) {
			test.Errorf("Case %d (%+v): Incorrect Output. Expected: '%v', Actual: '%v'.", i,
				testCase, testCase.expectedOutput, actualOutput)
			continue
		}

		logs, err := GetLogRecords(log.LevelWarn, time.Time{}, "", "", "")
		if err != nil {
			test.Errorf("Case %d (%+v): Error getting log records.", i, testCase)
			continue
		}

		if testCase.numWarnings != len(logs) {
			test.Errorf("Case %d (%+v): Incorrect number of warnings issued. Expected: %d, Actual: %d.", i,
				testCase, testCase.numWarnings, len(logs))
			continue
		}
	}
}
