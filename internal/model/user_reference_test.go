package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestResolveCourseUserReferences(test *testing.T) {
	testCases := []struct {
		input      []CourseUserReference
		output     *ResolvedCourseUserReference
		userErrors map[string]error
	}{
		// Target Emails
		{
			[]CourseUserReference{"course-student@test.edulinq.org"},
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			nil,
		},
		{
			[]CourseUserReference{"-course-student@test.edulinq.org"},
			&ResolvedCourseUserReference{
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			nil,
		},

		// Target Roles
		{
			[]CourseUserReference{"admin"},
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},
		{
			[]CourseUserReference{"-admin"},
			&ResolvedCourseUserReference{
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},

		// All Users
		{
			[]CourseUserReference{"*"},
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			nil,
		},
		{
			[]CourseUserReference{"-*"},
			&ResolvedCourseUserReference{
				ExcludeCourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			nil,
		},

		// Complex, Normalization
		{
			[]CourseUserReference{
				"course-student@test.edulinq.org",
				"COURSE-student@test.EDULINQ.org",
				"admin",
				"aDmIn",
			},
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},
		{
			[]CourseUserReference{
				"course-student@test.edulinq.org    	",
				"    	course-student@test.edulinq.org",
				"   -admin",
				"-admin	",
			},
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},

		// Complex, Non-Overlapping
		{
			[]CourseUserReference{
				"course-student@test.edulinq.org",
				"-course-admin@test.edulinq.org",
				"admin",
				"-owner",
			},
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"owner": nil,
				},
			},
			nil,
		},

		// Complex, Overlapping
		{
			[]CourseUserReference{
				"course-student@test.edulinq.org",
				"-course-student@test.edulinq.org",
				"admin",
				"-admin",
			},
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},
		{
			[]CourseUserReference{
				"grader",
				"*",
			},
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			nil,
		},
		{
			[]CourseUserReference{
				"-*",
				"-student",
			},
			&ResolvedCourseUserReference{
				ExcludeCourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			nil,
		},

		// Not Enrolled Users
		{
			[]CourseUserReference{"zzz@test.edulinq.org"},
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"zzz@test.edulinq.org": nil,
				},
			},
			nil,
		},
		{
			[]CourseUserReference{"server-user@test.edulinq.org"},
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"server-user@test.edulinq.org": nil,
				},
			},
			nil,
		},

		// Errors

		// Unknown Course Role
		{
			[]CourseUserReference{"ZZZ"},
			nil,
			map[string]error{
				"ZZZ": fmt.Errorf("Unknown course role 'zzz'."),
			},
		},
	}

	for i, testCase := range testCases {
		result, userErrors := ResolveCourseUserReferences(testCase.input)
		if !reflect.DeepEqual(userErrors, testCase.userErrors) {
			test.Errorf("Case %d: Unexpected user errors. Expected: '%v', Actual: '%v'.",
				i, testCase.userErrors, userErrors)
			continue
		}

		if len(testCase.userErrors) != 0 {
			continue
		}

		// Set empty fields to pass equality check.
		if testCase.output.Emails == nil {
			testCase.output.Emails = make(map[string]any, 0)
		}

		if testCase.output.ExcludeEmails == nil {
			testCase.output.ExcludeEmails = make(map[string]any, 0)
		}

		if testCase.output.CourseUserRoles == nil {
			testCase.output.CourseUserRoles = make(map[string]any, 0)
		}

		if testCase.output.ExcludeCourseUserRoles == nil {
			testCase.output.ExcludeCourseUserRoles = make(map[string]any, 0)
		}

		if !reflect.DeepEqual(testCase.output, result) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.output), util.MustToJSONIndent(result))
			continue
		}
	}
}

func TestResolveCourseUserEmails(test *testing.T) {
	defaultUsers := map[string]*CourseUser{
		"course-admin@test.edulinq.org": &CourseUser{
			Email: "course-admin@test.edulinq.org",
			Role:  CourseRoleAdmin,
		},
		"course-grader@test.edulinq.org": &CourseUser{
			Email: "course-grader@test.edulinq.org",
			Role:  CourseRoleGrader,
		},
		"course-other@test.edulinq.org": &CourseUser{
			Email: "course-other@test.edulinq.org",
			Role:  CourseRoleOther,
		},
		"course-owner@test.edulinq.org": &CourseUser{
			Email: "course-owner@test.edulinq.org",
			Role:  CourseRoleOwner,
		},
		"course-student@test.edulinq.org": &CourseUser{
			Email: "course-student@test.edulinq.org",
			Role:  CourseRoleStudent,
		},
	}

	testCases := []struct {
		reference      *ResolvedCourseUserReference
		users          map[string]*CourseUser
		expectedOutput []string
	}{
		// Empty Inputs
		{
			nil,
			nil,
			nil,
		},
		{
			&ResolvedCourseUserReference{},
			nil,
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			nil,
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			map[string]*CourseUser{},
			[]string{},
		},

		// Course Role
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org"},
		},

		// All Users
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org", "course-grader@test.edulinq.org", "course-other@test.edulinq.org", "course-owner@test.edulinq.org", "course-student@test.edulinq.org"},
		},

		// Role With Multiple Users
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"student": nil,
				},
			},
			map[string]*CourseUser{
				"a_student@test.edulinq.org": &CourseUser{
					Email: "a_student@test.edulinq.org",
					Role:  CourseRoleStudent,
				},
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
				"b_student@test.edulinq.org": &CourseUser{
					Email: "b_student@test.edulinq.org",
					Role:  CourseRoleStudent,
				},
			},
			[]string{"a_student@test.edulinq.org", "b_student@test.edulinq.org", "course-student@test.edulinq.org"},
		},

		// Role With No Users
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"owner": nil,
				},
			},
			map[string]*CourseUser{
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
			},
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"owner":   nil,
					"student": nil,
				},
			},
			map[string]*CourseUser{
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
			},
			[]string{"course-student@test.edulinq.org"},
		},

		// Exclude Email
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{
				"course-grader@test.edulinq.org",
				"course-other@test.edulinq.org",
				"course-owner@test.edulinq.org",
				"course-student@test.edulinq.org",
			},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org":   nil,
					"course-other@test.edulinq.org":   nil,
					"course-student@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{"course-grader@test.edulinq.org", "course-owner@test.edulinq.org"},
		},
		{
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-other@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org"},
		},
		{
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{},
		},

		// Exclude Role
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"student": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{"course-student@test.edulinq.org"},
		},
		{
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{"course-student@test.edulinq.org"},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{
				"course-grader@test.edulinq.org",
				"course-other@test.edulinq.org",
				"course-owner@test.edulinq.org",
				"course-student@test.edulinq.org",
			},
		},
	}

	for i, testCase := range testCases {
		actualOutput := ResolveCourseUserEmails(testCase.users, testCase.reference)

		if !reflect.DeepEqual(testCase.expectedOutput, actualOutput) {
			test.Errorf("Case %d: Incorrect Output. Expected: '%v', Actual: '%v'.",
				i, util.MustToJSONIndent(testCase.expectedOutput), util.MustToJSONIndent(actualOutput))
			continue
		}
	}
}

func TestResolveCourseUsers(test *testing.T) {
	defaultUsers := map[string]*CourseUser{
		"course-admin@test.edulinq.org": &CourseUser{
			Email: "course-admin@test.edulinq.org",
			Role:  CourseRoleAdmin,
		},
		"course-grader@test.edulinq.org": &CourseUser{
			Email: "course-grader@test.edulinq.org",
			Role:  CourseRoleGrader,
		},
		"course-other@test.edulinq.org": &CourseUser{
			Email: "course-owner@test.edulinq.org",
			Role:  CourseRoleOther,
		},
		"course-owner@test.edulinq.org": &CourseUser{
			Email: "course-owner@test.edulinq.org",
			Role:  CourseRoleOwner,
		},
		"course-student@test.edulinq.org": &CourseUser{
			Email: "course-student@test.edulinq.org",
			Role:  CourseRoleStudent,
		},
	}

	testCases := []struct {
		reference      *ResolvedCourseUserReference
		users          map[string]*CourseUser
		expectedOutput []string
	}{
		// Empty Inputs
		{
			nil,
			nil,
			nil,
		},
		{
			&ResolvedCourseUserReference{},
			nil,
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			nil,
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			map[string]*CourseUser{},
			[]string{},
		},

		// Course Role
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org"},
		},

		// All Users
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org", "course-grader@test.edulinq.org", "course-other@test.edulinq.org", "course-owner@test.edulinq.org", "course-student@test.edulinq.org"},
		},

		// Role With Multiple Users
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"student": nil,
				},
			},
			map[string]*CourseUser{
				"a_student@test.edulinq.org": &CourseUser{
					Email: "a_student@test.edulinq.org",
					Role:  CourseRoleStudent,
				},
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
				"b_student@test.edulinq.org": &CourseUser{
					Email: "b_student@test.edulinq.org",
					Role:  CourseRoleStudent,
				},
			},
			[]string{"a_student@test.edulinq.org", "b_student@test.edulinq.org", "course-student@test.edulinq.org"},
		},

		// Role With No Users
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"owner": nil,
				},
			},
			map[string]*CourseUser{
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
			},
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"owner":   nil,
					"student": nil,
				},
			},
			map[string]*CourseUser{
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
			},
			[]string{"course-student@test.edulinq.org"},
		},

		// Exclude Email
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{
				"course-grader@test.edulinq.org",
				"course-other@test.edulinq.org",
				"course-owner@test.edulinq.org",
				"course-student@test.edulinq.org",
			},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org":   nil,
					"course-other@test.edulinq.org":   nil,
					"course-student@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{"course-grader@test.edulinq.org", "course-owner@test.edulinq.org"},
		},
		{
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-other@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org"},
		},
		{
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{},
		},

		// Exclude Role
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: map[string]any{
					"student": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{"course-student@test.edulinq.org"},
		},
		{
			&ResolvedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{"course-student@test.edulinq.org"},
		},
		{
			&ResolvedCourseUserReference{
				CourseUserRoles: GetCommonCourseUserRoleStrings(),
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			defaultUsers,
			[]string{
				"course-grader@test.edulinq.org",
				"course-other@test.edulinq.org",
				"course-owner@test.edulinq.org",
				"course-student@test.edulinq.org",
			},
		},
	}

	for i, testCase := range testCases {
		actualOutput := ResolveCourseUsers(testCase.users, testCase.reference)

		expectedOutput := []*CourseUser{}

		for _, email := range testCase.expectedOutput {
			user, ok := testCase.users[email]
			if !ok {
				test.Errorf("Case %d: Failed to get expected course user: '%s'.", i, email)
				continue
			}

			expectedOutput = append(expectedOutput, user)
		}

		if testCase.expectedOutput == nil {
			expectedOutput = nil
		}

		if !reflect.DeepEqual(expectedOutput, actualOutput) {
			test.Errorf("Case %d: Incorrect Output. Expected: '%v', Actual: '%v'.",
				i, util.MustToJSONIndent(expectedOutput), util.MustToJSONIndent(actualOutput))
			continue
		}
	}
}
