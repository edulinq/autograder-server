package db

import (
	"reflect"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

func (this *DBTests) DBTestResolveCourseUserEmails(test *testing.T) {
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
		addUsers       map[string]*model.ServerUser
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
			[]string{"b@test.edulinq.org", "a@test.edulinq.org", "c@test.edulinq.org"},
			[]string{"a@test.edulinq.org", "b@test.edulinq.org", "c@test.edulinq.org"},
			nil,
			[]string{},
			0,
		},

		// This is a test to ensure miscapitalized emails only get returned once.
		{
			[]string{"a@test.edulinq.org", "A@tesT.EduLiNq.OrG", "A@TESt.EDuLINQ.ORG"},
			[]string{"a@test.edulinq.org"},
			nil,
			[]string{},
			0,
		},

		// This is a basic test to ensure that a role gets mapped to the correct email.
		{
			[]string{"admin"},
			[]string{"course-admin@test.edulinq.org"},
			nil,
			[]string{},
			0,
		},

		// This is a test for our all roles character, the *.
		{
			[]string{"*"},
			[]string{"course-admin@test.edulinq.org", "course-grader@test.edulinq.org", "course-other@test.edulinq.org", "course-owner@test.edulinq.org", "course-student@test.edulinq.org"},
			nil,
			[]string{},
			0,
		},

		// This test case is given redundant roles and emails.
		// It tests to ensures we do not produce duplicates on this input.
		{
			[]string{"other", "*", "course-grader@test.edulinq.org"},
			[]string{"course-admin@test.edulinq.org", "course-grader@test.edulinq.org", "course-other@test.edulinq.org", "course-owner@test.edulinq.org", "course-student@test.edulinq.org"},
			nil,
			[]string{},
			0,
		},

		// This test case tests if miscapitalized roles still function.
		{
			[]string{"OTHER"},
			[]string{"course-other@test.edulinq.org"},
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
			[]string{"a_student@test.edulinq.org", "b_student@test.edulinq.org", "course-student@test.edulinq.org"},
			map[string]*model.ServerUser{
				"a_student@test.edulinq.org": &model.ServerUser{
					Email: "a_student@test.edulinq.org",
					Name:  nil,
					Role:  model.ServerRoleUser,
					CourseInfo: map[string]*model.UserCourseInfo{
						TEST_COURSE_ID: &model.UserCourseInfo{
							Role:  model.CourseRoleStudent,
							LMSID: nil,
						},
					},
				},
				"b_student@test.edulinq.org": &model.ServerUser{
					Email: "b_student@test.edulinq.org",
					Name:  nil,
					Role:  model.ServerRoleUser,
					CourseInfo: map[string]*model.UserCourseInfo{
						TEST_COURSE_ID: &model.UserCourseInfo{
							Role:  model.CourseRoleStudent,
							LMSID: nil,
						},
					},
				},
			},
			[]string{},
			0,
		},

		// This is a test case to see if we properly trim whitespace.
		{
			[]string{"\t\n student    ", "\n \t testing@test.edulinq.org", "\t\n     \t    \n"},
			[]string{"course-student@test.edulinq.org", "testing@test.edulinq.org"},
			nil,
			[]string{},
			0,
		},

		// This test case removes the only user from the "owner" role, so we check that a role without any users still functions properly.
		{
			[]string{"owner", "student"},
			[]string{"course-student@test.edulinq.org"},
			nil,
			[]string{"course-owner@test.edulinq.org"},
			0,
		},

		// This test supplies a single role that resolves to nothing.
		{
			[]string{"owner"},
			[]string{},
			nil,
			[]string{"course-owner@test.edulinq.org"},
			0,
		},

		// Remove a user.
		{
			[]string{"*", "-course-admin@test.edulinq.org"},
			[]string{"course-grader@test.edulinq.org", "course-other@test.edulinq.org", "course-owner@test.edulinq.org", "course-student@test.edulinq.org"},
			nil,
			[]string{},
			0,
		},

		// Remove multiple users.
		{
			[]string{"*", "-course-admin@test.edulinq.org", "- course-other@test.edulinq.org", " - course-student@test.edulinq.org"},
			[]string{"course-grader@test.edulinq.org", "course-owner@test.edulinq.org"},
			nil,
			[]string{},
			0,
		},

		// Remove a use that is not in the list.
		{
			[]string{"course-admin@test.edulinq.org", "-course-other@test.edulinq.org"},
			[]string{"course-admin@test.edulinq.org"},
			nil,
			[]string{},
			0,
		},

		// Remove all users.
		{
			[]string{"course-admin@test.edulinq.org", "-course-admin@test.edulinq.org"},
			[]string{},
			nil,
			[]string{},
			0,
		},
	}

	for i, testCase := range testCases {
		ResetForTesting()
		course := MustGetCourse(TEST_COURSE_ID)

		if testCase.addUsers != nil {
			UpsertUsers(testCase.addUsers)
		}

		for _, removeUser := range testCase.removeUsers {
			RemoveUserFromCourse(course, removeUser)
		}

		actualOutput, err := ResolveCourseUserEmails(course, testCase.input)
		if err != nil {
			test.Errorf("Case %d (%+v): Resolve user failed: '%v'.", i, testCase, err)
			continue
		}

		if !reflect.DeepEqual(testCase.expectedOutput, actualOutput) {
			test.Errorf("Case %d (%+v): Incorrect Output. Expected: '%v', Actual: '%v'.", i,
				testCase, testCase.expectedOutput, actualOutput)
			continue
		}

		logs, err := GetLogRecords(log.ParsedLogQuery{Level: log.LevelWarn})
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
