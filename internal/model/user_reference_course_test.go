package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

// All common course user roles that can be placed in a ParsedCourseUserReference.
var allCourseRoles = getAllCourseRoles()

// The test course users found in the test data.
// Explicitly created to avoid an import cycle with the db.
var defaultCourseUsers = map[string]*CourseUser{
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

// The shared test cases between the TestResolveCourseUser*() functions.
var resolveCourseUserTestCases = []resolveCourseUserTestCase{
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
		defaultCourseUsers,
		[]string{"course-admin@test.edulinq.org"},
	},

	// All Users
	{
		&ParsedCourseUserReference{
			CourseUserRoles: allCourseRoles,
		},
		defaultCourseUsers,
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
			"course-student@test.edulinq.org": defaultCourseUsers["course-student@test.edulinq.org"],
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
			"course-student@test.edulinq.org": defaultCourseUsers["course-student@test.edulinq.org"],
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
			"course-student@test.edulinq.org": defaultCourseUsers["course-student@test.edulinq.org"],
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
		defaultCourseUsers,
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
		defaultCourseUsers,
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
		defaultCourseUsers,
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
		defaultCourseUsers,
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
		defaultCourseUsers,
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
		defaultCourseUsers,
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
		defaultCourseUsers,
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
		defaultCourseUsers,
		[]string{"course-student@test.edulinq.org"},
	},
	{
		&ParsedCourseUserReference{
			CourseUserRoles: allCourseRoles,
			ExcludeCourseUserRoles: map[CourseUserRole]any{
				GetCourseUserRole("admin"): nil,
			},
		},
		defaultCourseUsers,
		[]string{
			"course-grader@test.edulinq.org",
			"course-other@test.edulinq.org",
			"course-owner@test.edulinq.org",
			"course-student@test.edulinq.org",
		},
	},
}

// The named test case struct allows specific tests to add additional test cases.
type resolveCourseUserTestCase struct {
	reference      *ParsedCourseUserReference
	users          map[string]*CourseUser
	expectedOutput []string
}

func TestParseCourseUserReferences(test *testing.T) {
	testCases := []struct {
		input          []CourseUserReference
		output         *ParsedCourseUserReference
		errorSubstring string
	}{
		// Target Emails
		{
			[]CourseUserReference{"course-student@test.edulinq.org"},
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			"",
		},
		{
			[]CourseUserReference{"-course-student@test.edulinq.org"},
			&ParsedCourseUserReference{
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			"",
		},

		// Target Roles
		{
			[]CourseUserReference{"admin"},
			&ParsedCourseUserReference{
				CourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			"",
		},
		{
			[]CourseUserReference{"-admin"},
			&ParsedCourseUserReference{
				ExcludeCourseUserRoles: map[CourseUserRole]any{
					GetCourseUserRole("admin"): nil,
				},
			},
			"",
		},

		// All Users
		{
			[]CourseUserReference{"*"},
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			"",
		},
		{
			[]CourseUserReference{"-*"},
			&ParsedCourseUserReference{
				ExcludeCourseUserRoles: allCourseRoles,
			},
			"",
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
			"",
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
			"",
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
			"",
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
			"",
		},
		{
			[]CourseUserReference{
				"grader",
				"*",
			},
			&ParsedCourseUserReference{
				CourseUserRoles: allCourseRoles,
			},
			"",
		},
		{
			[]CourseUserReference{
				"-*",
				"-student",
			},
			&ParsedCourseUserReference{
				ExcludeCourseUserRoles: allCourseRoles,
			},
			"",
		},

		// Not Enrolled Users
		{
			[]CourseUserReference{"zzz@test.edulinq.org"},
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"zzz@test.edulinq.org": nil,
				},
			},
			"",
		},
		{
			[]CourseUserReference{"server-user@test.edulinq.org"},
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"server-user@test.edulinq.org": nil,
				},
			},
			"",
		},

		// Errors

		// Unknown Course Role
		{
			[]CourseUserReference{"ZZZ"},
			nil,
			"Course user reference 'ZZZ' contains an unknown course role: 'zzz'.",
		},
	}

	for i, testCase := range testCases {
		result, err := ParseCourseUserReferences(testCase.input)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring: '%s', Actual Error: '%s'.", i, testCase.errorSubstring, err.Error())
				}
			} else {
				test.Errorf("Case %d: Failed to parse '%s': '%s'.", i, util.MustToJSONIndent(testCase.input), err.Error())
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error '%s'.", i, testCase.errorSubstring)
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
	extraTestCases := []resolveCourseUserTestCase{
		// Outside Emails
		{
			&ParsedCourseUserReference{
				Emails: map[string]any{
					"outside-email@test.edulinq.org": nil,
				},
			},
			defaultCourseUsers,
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
			defaultCourseUsers,
			[]string{},
		},
	}

	testCases := make([]resolveCourseUserTestCase, len(resolveCourseUserTestCases))
	copy(resolveCourseUserTestCases, testCases)

	testCases = append(testCases, extraTestCases...)

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
	for i, testCase := range resolveCourseUserTestCases {
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

func getAllCourseRoles() map[CourseUserRole]any {
	commonCourseRoles := GetCommonCourseUserRolesCopy()

	allCourseRoles := make(map[CourseUserRole]any, len(commonCourseRoles))
	for _, role := range commonCourseRoles {
		allCourseRoles[role] = nil
	}

	return allCourseRoles
}
