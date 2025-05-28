package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestParseCourseUserReferences(test *testing.T) {
	commonCourseRoles := GetCommonCourseUserRolesCopy()

	allCourseRoles := make(map[CourseUserRole]any, len(commonCourseRoles))
	for _, role := range commonCourseRoles {
		allCourseRoles[role] = nil
	}

	testCases := []struct {
		input      []CourseUserReference
		output     *ParsedCourseUserReference
		userErrors map[string]error
	}{
		// Target Emails
		{
			[]CourseUserReference{"course-student@test.edulinq.org"},
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			nil,
		},
		{
			[]CourseUserReference{"-course-student@test.edulinq.org"},
			&ParsedCourseUserReference{
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			nil,
		},

		// Target Roles
		{
			[]CourseUserReference{"admin"},
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			nil,
		},
		{
			[]CourseUserReference{"-admin"},
			&ParsedCourseUserReference{
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			nil,
		},

		// All Users
		{
			[]CourseUserReference{"*"},
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			nil,
		},
		{
			[]CourseUserReference{"-*"},
			&ParsedCourseUserReference{
				ExcludeCourseUserRoles: allCourseRoles,
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
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
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
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
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
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("owner"): nil,
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
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			nil,
		},
		{
			[]CourseUserReference{
				"grader",
				"*",
			},
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			nil,
		},
		{
			[]CourseUserReference{
				"-*",
				"-student",
			},
			&ParsedCourseUserReference{
				ExcludeCourseUserRoles: allCourseRoles,
			},
			nil,
		},

		// Not Enrolled Users
		{
			[]CourseUserReference{"zzz@test.edulinq.org"},
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"zzz@test.edulinq.org": nil,
				},
			},
			nil,
		},
		{
			[]CourseUserReference{"server-user@test.edulinq.org"},
			&ParsedCourseUserReference{
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
		result, userErrors := ParseCourseUserReferences(testCase.input)
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
			testCase.output.CourseUserRoles = make(map[CourseUserRole]any, 0)
		}

		if testCase.output.ExcludeCourseUserRoles == nil {
			testCase.output.ExcludeCourseUserRoles = make(map[CourseUserRole]any, 0)
		}

		if !reflect.DeepEqual(testCase.output, result) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.output), util.MustToJSONIndent(result))
			continue
		}
	}
}

func TestResolveCourseUserEmails(test *testing.T) {
	commonCourseRoles := GetCommonCourseUserRolesCopy()

	allCourseRoles := make(map[CourseUserRole]any, len(commonCourseRoles))
	for _, role := range commonCourseRoles {
		allCourseRoles[role] = nil
	}

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
		reference      *ParsedCourseUserReference
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
			&ParsedCourseUserReference{},
			nil,
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			nil,
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			map[string]*CourseUser{},
			[]string{},
		},

		// Course Role
		{
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org"},
		},

		// All Users
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org", "course-grader@test.edulinq.org", "course-other@test.edulinq.org", "course-owner@test.edulinq.org", "course-student@test.edulinq.org"},
		},

		// Role With Multiple Users
		{
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("student"): nil,
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
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("owner"): nil,
				},
			},
			map[string]*CourseUser{
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
			},
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("owner"):   nil,
					GetCourseUserRole("student"): nil,
				},
			},
			map[string]*CourseUser{
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
			},
			[]string{"course-student@test.edulinq.org"},
		},

		// Exclude Email
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
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
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
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
			&ParsedCourseUserReference{
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
			&ParsedCourseUserReference{
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
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("student"): nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{"course-student@test.edulinq.org"},
		},
		{
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{"course-student@test.edulinq.org"},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
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

		// Outside Emails
		{
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"outside-email@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{"outside-email@test.edulinq.org"},
		},
		{
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"outside-email@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"outside-email@test.edulinq.org": nil,
				},
			},
			defaultUsers,
			[]string{},
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
	commonCourseRoles := GetCommonCourseUserRolesCopy()

	allCourseRoles := make(map[CourseUserRole]any, len(commonCourseRoles))
	for _, role := range commonCourseRoles {
		allCourseRoles[role] = nil
	}

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
		reference      *ParsedCourseUserReference
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
			&ParsedCourseUserReference{},
			nil,
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			nil,
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			map[string]*CourseUser{},
			[]string{},
		},

		// Course Role
		{
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org"},
		},

		// All Users
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			defaultUsers,
			[]string{"course-admin@test.edulinq.org", "course-grader@test.edulinq.org", "course-other@test.edulinq.org", "course-owner@test.edulinq.org", "course-student@test.edulinq.org"},
		},

		// Role With Multiple Users
		{
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("student"): nil,
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
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("owner"): nil,
				},
			},
			map[string]*CourseUser{
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
			},
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("owner"):   nil,
					GetCourseUserRole("student"): nil,
				},
			},
			map[string]*CourseUser{
				"course-student@test.edulinq.org": defaultUsers["course-student@test.edulinq.org"],
			},
			[]string{"course-student@test.edulinq.org"},
		},

		// Exclude Email
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
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
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
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
			&ParsedCourseUserReference{
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
			&ParsedCourseUserReference{
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
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("student"): nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{"course-student@test.edulinq.org"},
		},
		{
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			defaultUsers,
			[]string{"course-student@test.edulinq.org"},
		},
		{
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
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
