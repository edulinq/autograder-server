package model

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

// All common server user roles that can be placed in a ParsedServerUserReference.
var allServerRoles = getAllServerRoles()

// The test courses found in the test data.
// Explicitly created to avoid an import cycle with the db.
var testCourses = map[string]*Course{
	TEST_COURSE_ID: &Course{
		ID: TEST_COURSE_ID,
	},
	"course-languages": &Course{
		ID: "course-languages",
	},
}

// The test server users found in the test data.
// Explicitly created to avoid an import cycle with the db.
var defaultServerUsers = map[string]*ServerUser{
	"course-admin@test.edulinq.org": &ServerUser{
		Email: "course-admin@test.edulinq.org",
		Role:  ServerRoleUser,
		CourseInfo: map[string]*UserCourseInfo{
			TEST_COURSE_ID: &UserCourseInfo{
				Role: CourseRoleAdmin,
			},
			"course-languages": &UserCourseInfo{
				Role: CourseRoleAdmin,
			},
		},
	},
	"course-grader@test.edulinq.org": &ServerUser{
		Email: "course-grader@test.edulinq.org",
		Role:  ServerRoleUser,
		CourseInfo: map[string]*UserCourseInfo{
			TEST_COURSE_ID: &UserCourseInfo{
				Role: CourseRoleGrader,
			},
			"course-languages": &UserCourseInfo{
				Role: CourseRoleGrader,
			},
		},
	},
	"course-other@test.edulinq.org": &ServerUser{
		Email: "course-other@test.edulinq.org",
		Role:  ServerRoleUser,
		CourseInfo: map[string]*UserCourseInfo{
			TEST_COURSE_ID: &UserCourseInfo{
				Role: CourseRoleOther,
			},
			"course-languages": &UserCourseInfo{
				Role: CourseRoleOther,
			},
		},
	},
	"course-owner@test.edulinq.org": &ServerUser{
		Email: "course-owner@test.edulinq.org",
		Role:  ServerRoleUser,
		CourseInfo: map[string]*UserCourseInfo{
			TEST_COURSE_ID: &UserCourseInfo{
				Role: CourseRoleOwner,
			},
			"course-languages": &UserCourseInfo{
				Role: CourseRoleOwner,
			},
		},
	},
	"course-student@test.edulinq.org": &ServerUser{
		Email: "course-student@test.edulinq.org",
		Role:  ServerRoleUser,
		CourseInfo: map[string]*UserCourseInfo{
			TEST_COURSE_ID: &UserCourseInfo{
				Role: CourseRoleStudent,
			},
			"course-languages": &UserCourseInfo{
				Role: CourseRoleStudent,
			},
		},
	},
	"server-admin@test.edulinq.org": &ServerUser{
		Email:      "server-admin@test.edulinq.org",
		Role:       ServerRoleAdmin,
		CourseInfo: nil,
	},
	"server-creator@test.edulinq.org": &ServerUser{
		Email:      "server-creator@test.edulinq.org",
		Role:       ServerRoleCourseCreator,
		CourseInfo: nil,
	},
	"server-owner@test.edulinq.org": &ServerUser{
		Email:      "server-owner@test.edulinq.org",
		Role:       ServerRoleOwner,
		CourseInfo: nil,
	},
	"server-user@test.edulinq.org": &ServerUser{
		Email:      "server-user@test.edulinq.org",
		Role:       ServerRoleUser,
		CourseInfo: nil,
	},
}

// The shared test cases between the TestResolveServerUser*() functions.
var resolveServerUserTestCases = []resolveServerUserTestCase{
	// Empty Inputs
	{
		nil,
		nil,
		nil,
	},
	{
		&ParsedServerUserReference{},
		nil,
		[]string{},
	},
	{
		&ParsedServerUserReference{
			ServerUserRoles: allServerRoles,
		},
		nil,
		[]string{},
	},
	{
		&ParsedServerUserReference{
			ServerUserRoles: allServerRoles,
		},
		map[string]*ServerUser{},
		[]string{},
	},

	// Server Role
	{
		&ParsedServerUserReference{
			ServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("admin"): nil,
			},
		},
		defaultServerUsers,
		[]string{"server-admin@test.edulinq.org"},
	},

	// All Users
	{
		&ParsedServerUserReference{
			ServerUserRoles: allServerRoles,
		},
		defaultServerUsers,
		[]string{
			"course-admin@test.edulinq.org",
			"course-grader@test.edulinq.org",
			"course-other@test.edulinq.org",
			"course-owner@test.edulinq.org",
			"course-student@test.edulinq.org",
			"server-admin@test.edulinq.org",
			"server-creator@test.edulinq.org",
			"server-owner@test.edulinq.org",
			"server-user@test.edulinq.org",
		},
	},

	// Server Role With Multiple Users
	{
		&ParsedServerUserReference{
			ServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("user"): nil,
			},
		},
		defaultServerUsers,
		[]string{
			"course-admin@test.edulinq.org",
			"course-grader@test.edulinq.org",
			"course-other@test.edulinq.org",
			"course-owner@test.edulinq.org",
			"course-student@test.edulinq.org",
			"server-user@test.edulinq.org",
		},
	},

	// Server Role With No Users
	{
		&ParsedServerUserReference{
			ServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("owner"): nil,
			},
		},
		map[string]*ServerUser{
			"server-creator@test.edulinq.org": defaultServerUsers["server-creator@test.edulinq.org"],
		},
		[]string{},
	},
	{
		&ParsedServerUserReference{
			ServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("creator"): nil,
				GetServerUserRole("owner"):   nil,
			},
		},
		map[string]*ServerUser{
			"server-creator@test.edulinq.org": defaultServerUsers["server-creator@test.edulinq.org"],
		},
		[]string{"server-creator@test.edulinq.org"},
	},

	// Exclude Email
	{
		&ParsedServerUserReference{
			ServerUserRoles: allServerRoles,
			ExcludeEmails: map[string]any{
				"course-other@test.edulinq.org":   nil,
				"course-student@test.edulinq.org": nil,
				"server-admin@test.edulinq.org":   nil,
			},
		},
		defaultServerUsers,
		[]string{
			"course-admin@test.edulinq.org",
			"course-grader@test.edulinq.org",
			"course-owner@test.edulinq.org",
			"server-creator@test.edulinq.org",
			"server-owner@test.edulinq.org",
			"server-user@test.edulinq.org",
		},
	},
	{
		&ParsedServerUserReference{
			Emails: map[string]any{
				"server-admin@test.edulinq.org": nil,
			},
			ExcludeEmails: map[string]any{
				"server-other@test.edulinq.org": nil,
			},
		},
		defaultServerUsers,
		[]string{"server-admin@test.edulinq.org"},
	},
	{
		&ParsedServerUserReference{
			Emails: map[string]any{
				"server-admin@test.edulinq.org": nil,
			},
			ExcludeEmails: map[string]any{
				"server-admin@test.edulinq.org": nil,
			},
		},
		defaultServerUsers,
		[]string{},
	},

	// Exclude Role
	{
		&ParsedServerUserReference{
			ServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("admin"): nil,
			},
			ExcludeServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("admin"): nil,
			},
		},
		defaultServerUsers,
		[]string{},
	},
	{
		&ParsedServerUserReference{
			Emails: map[string]any{
				"server-admin@test.edulinq.org": nil,
			},
			ExcludeServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("admin"): nil,
			},
		},
		defaultServerUsers,
		[]string{},
	},
	{
		&ParsedServerUserReference{
			ServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("owner"): nil,
			},
			ExcludeServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("admin"): nil,
			},
		},
		defaultServerUsers,
		[]string{"server-owner@test.edulinq.org"},
	},
	{
		&ParsedServerUserReference{
			Emails: map[string]any{
				"server-user@test.edulinq.org": nil,
			},
			ExcludeServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("admin"): nil,
			},
		},
		defaultServerUsers,
		[]string{"server-user@test.edulinq.org"},
	},
	{
		&ParsedServerUserReference{
			ServerUserRoles: allServerRoles,
			ExcludeServerUserRoles: map[ServerUserRole]any{
				GetServerUserRole("user"): nil,
			},
		},
		defaultServerUsers,
		[]string{
			"server-admin@test.edulinq.org",
			"server-creator@test.edulinq.org",
			"server-owner@test.edulinq.org",
		},
	},

	// Course Role
	{
		&ParsedServerUserReference{
			CourseUserReferences: map[string]*ParsedCourseUserReference{
				TEST_COURSE_ID: &ParsedCourseUserReference{
					CourseUserRoles: map[CourseUserRole]any{
						GetCourseUserRole("admin"): nil,
					},
				},
			},
		},
		defaultServerUsers,
		[]string{
			"course-admin@test.edulinq.org",
		},
	},

	// All Course Roles
	{
		&ParsedServerUserReference{
			CourseUserReferences: map[string]*ParsedCourseUserReference{
				TEST_COURSE_ID: &ParsedCourseUserReference{
					CourseUserRoles: allCourseRoles,
				},
			},
		},
		defaultServerUsers,
		[]string{
			"course-admin@test.edulinq.org",
			"course-grader@test.edulinq.org",
			"course-other@test.edulinq.org",
			"course-owner@test.edulinq.org",
			"course-student@test.edulinq.org",
		},
	},

	// Exclude Course Roles
	{
		&ParsedServerUserReference{
			ServerUserRoles: allServerRoles,
			CourseUserReferences: map[string]*ParsedCourseUserReference{
				TEST_COURSE_ID: &ParsedCourseUserReference{
					ExcludeCourseUserRoles: allCourseRoles,
				},
			},
		},
		defaultServerUsers,
		[]string{
			"server-admin@test.edulinq.org",
			"server-creator@test.edulinq.org",
			"server-owner@test.edulinq.org",
			"server-user@test.edulinq.org",
		},
	},

	// Conflicts Between Courses
	{
		&ParsedServerUserReference{
			CourseUserReferences: map[string]*ParsedCourseUserReference{
				TEST_COURSE_ID: &ParsedCourseUserReference{
					CourseUserRoles: allCourseRoles,
				},
				"course-languages": &ParsedCourseUserReference{
					ExcludeEmails: map[string]any{
						"course-admin@test.edulinq.org": nil,
					},
				},
			},
		},
		defaultServerUsers,
		[]string{
			"course-grader@test.edulinq.org",
			"course-other@test.edulinq.org",
			"course-owner@test.edulinq.org",
			"course-student@test.edulinq.org",
		},
	},
	{
		&ParsedServerUserReference{
			CourseUserReferences: map[string]*ParsedCourseUserReference{
				TEST_COURSE_ID: &ParsedCourseUserReference{
					CourseUserRoles: allCourseRoles,
				},
				"course-languages": &ParsedCourseUserReference{
					ExcludeCourseUserRoles: allCourseRoles,
				},
			},
		},
		defaultServerUsers,
		[]string{},
	},
}

// The named test case struct allows specific tests to add additional test cases.
type resolveServerUserTestCase struct {
	reference      *ParsedServerUserReference
	users          map[string]*ServerUser
	expectedOutput []string
}

func TestParseServerUserReferences(test *testing.T) {
	testCases := []struct {
		input          []ServerUserReference
		output         *ParsedServerUserReference
		errorSubstring string
	}{
		// All Users
		{
			[]ServerUserReference{"*"},
			&ParsedServerUserReference{
				ServerUserRoles: allServerRoles,
			},
			"",
		},
		{
			[]ServerUserReference{"-*"},
			&ParsedServerUserReference{
				ExcludeServerUserRoles: allServerRoles,
			},
			"",
		},

		// Target Email
		{
			[]ServerUserReference{"course-student@test.edulinq.org"},
			&ParsedServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			"",
		},
		{
			[]ServerUserReference{"-course-student@test.edulinq.org"},
			&ParsedServerUserReference{
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			"",
		},

		// Target Server Role
		{
			[]ServerUserReference{"user"},
			&ParsedServerUserReference{
				ServerUserRoles: map[ServerUserRole]any{
					GetServerUserRole("user"): nil,
				},
			},
			"",
		},
		{
			[]ServerUserReference{"-user"},
			&ParsedServerUserReference{
				ExcludeServerUserRoles: map[ServerUserRole]any{
					GetServerUserRole("user"): nil,
				},
			},
			"",
		},

		// All Courses, All Course Roles
		{
			[]ServerUserReference{"*::*"},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: allCourseRoles,
					},
					"course-languages": &ParsedCourseUserReference{
						CourseUserRoles: allCourseRoles,
					},
				},
			},
			"",
		},
		{
			[]ServerUserReference{"-*::*"},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						ExcludeCourseUserRoles: allCourseRoles,
					},
					"course-languages": &ParsedCourseUserReference{
						ExcludeCourseUserRoles: allCourseRoles,
					},
				},
			},
			"",
		},

		// All Courses, Target Course Role
		{
			[]ServerUserReference{"*::student"},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("student"): nil,
						},
					},
					"course-languages": &ParsedCourseUserReference{
						CourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("student"): nil,
						},
					},
				},
			},
			"",
		},
		{
			[]ServerUserReference{"-*::student"},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						ExcludeCourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("student"): nil,
						},
					},
					"course-languages": &ParsedCourseUserReference{
						ExcludeCourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("student"): nil,
						},
					},
				},
			},
			"",
		},

		// Target Course, All Course Roles
		{
			[]ServerUserReference{
				ServerUserReference(fmt.Sprintf("%s::*", TEST_COURSE_ID)),
			},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: allCourseRoles,
					},
				},
			},
			"",
		},
		{
			[]ServerUserReference{
				ServerUserReference(fmt.Sprintf("-%s::*", TEST_COURSE_ID)),
			},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						ExcludeCourseUserRoles: allCourseRoles,
					},
				},
			},
			"",
		},

		// Target Course, Target Course Role
		{
			[]ServerUserReference{
				ServerUserReference(fmt.Sprintf("%s::student", TEST_COURSE_ID)),
			},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("student"): nil,
						},
					},
				},
			},
			"",
		},
		{
			[]ServerUserReference{
				ServerUserReference(fmt.Sprintf("-%s::student", TEST_COURSE_ID)),
			},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						ExcludeCourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("student"): nil,
						},
					},
				},
			},
			"",
		},

		// Complex, Normalization
		{
			[]ServerUserReference{
				"course-student@test.edulinq.org",
				"COURSE-student@test.EDULINQ.org",
				"admin",
				"aDmIn",
				"COURSE101::grader",
				"course101::GRADER",
			},
			&ParsedServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ServerUserRoles: map[ServerUserRole]any{
					GetServerUserRole("admin"): nil,
				},
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("grader"): nil,
						},
					},
				},
			},
			"",
		},
		{
			[]ServerUserReference{
				"course-student@test.edulinq.org    	",
				"    	course-student@test.edulinq.org",
				"   admin",
				"admin	",
				"	course101     ::   grader	",
				" course101	::	grader     ",
			},
			&ParsedServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ServerUserRoles: map[ServerUserRole]any{
					GetServerUserRole("admin"): nil,
				},
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("grader"): nil,
						},
					},
				},
			},
			"",
		},

		// Complex, Non-Overlapping
		{
			[]ServerUserReference{
				"course-student@test.edulinq.org",
				"-course-admin@test.edulinq.org",
				"admin",
				"-owner",
				"course101::grader",
				"-course101::student",
			},
			&ParsedServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ServerUserRoles: map[ServerUserRole]any{
					GetServerUserRole("admin"): nil,
				},
				ExcludeServerUserRoles: map[ServerUserRole]any{
					GetServerUserRole("owner"): nil,
				},
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("grader"): nil,
						},
						ExcludeCourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("student"): nil,
						},
					},
				},
			},
			"",
		},

		// Complex, Overlapping
		{
			[]ServerUserReference{
				"course-student@test.edulinq.org",
				"-course-student@test.edulinq.org",
				"admin",
				"-admin",
				"course101::grader",
				"-course101::grader",
			},
			&ParsedServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ServerUserRoles: map[ServerUserRole]any{
					GetServerUserRole("admin"): nil,
				},
				ExcludeServerUserRoles: map[ServerUserRole]any{
					GetServerUserRole("admin"): nil,
				},
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("grader"): nil,
						},
						ExcludeCourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("grader"): nil,
						},
					},
				},
			},
			"",
		},
		{
			[]ServerUserReference{
				"course101::grader",
				"*::grader",
			},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("grader"): nil,
						},
					},
					"course-languages": &ParsedCourseUserReference{
						CourseUserRoles: map[CourseUserRole]any{
							GetCourseUserRole("grader"): nil,
						},
					},
				},
			},
			"",
		},
		{
			[]ServerUserReference{
				"course101::grader",
				"course101::*",
			},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: allCourseRoles,
					},
				},
			},
			"",
		},
		{
			[]ServerUserReference{
				"course101::grader",
				"course101::*",
				"*::grader",
				"*::*",
			},
			&ParsedServerUserReference{
				CourseUserReferences: map[string]*ParsedCourseUserReference{
					TEST_COURSE_ID: &ParsedCourseUserReference{
						CourseUserRoles: allCourseRoles,
					},
					"course-languages": &ParsedCourseUserReference{
						CourseUserRoles: allCourseRoles,
					},
				},
			},
			"",
		},
		{
			[]ServerUserReference{
				"admin",
				"*",
			},
			&ParsedServerUserReference{
				ServerUserRoles: allServerRoles,
			},
			"",
		},

		// Errors

		// Accessing Root
		{
			[]ServerUserReference{"root"},
			nil,
			"Server user reference 'root' contains an unknown server role: 'root'.",
		},
		{
			[]ServerUserReference{"-root"},
			nil,
			"Server user reference '-root' contains an unknown server role: 'root'.",
		},

		// Unknown Server Role
		{
			[]ServerUserReference{"ZZZ"},
			nil,
			"Server user reference 'ZZZ' contains an unknown server role: 'zzz'.",
		},

		// Unknown Course
		{
			[]ServerUserReference{"ZZZ::*"},
			nil,
			"Server user reference 'ZZZ::*' contains an unknown course: 'zzz'.",
		},

		// Unknown Course Role
		{
			[]ServerUserReference{"*::ZZZ"},
			nil,
			"Server user reference '*::ZZZ' contains an unknown course role: 'zzz'.",
		},

		// Invalid Format
		{
			[]ServerUserReference{"foo::bar::baz"},
			nil,
			"Invalid format in server user reference: 'foo::bar::baz'.",
		},
	}

	for i, testCase := range testCases {
		result, err := ParseServerUserReferences(testCase.input, testCourses)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output. Expected Substring '%s', Actual Error: '%s'.",
						i, testCase.errorSubstring, err.Error())
				}
			} else {
				test.Errorf("Case %d: Failed to parse user reference '%s': '%v'.",
					i, util.MustToJSONIndent(testCase.output), err.Error())
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error for input '%s'.",
				i, util.MustToJSONIndent(testCase.input))
			continue
		}

		// Set empty fields to pass equality check.
		if testCase.output.Emails == nil {
			testCase.output.Emails = make(map[string]any, 0)
		}

		if testCase.output.ExcludeEmails == nil {
			testCase.output.ExcludeEmails = make(map[string]any, 0)
		}

		if testCase.output.ServerUserRoles == nil {
			testCase.output.ServerUserRoles = make(map[ServerUserRole]any, 0)
		}

		if testCase.output.ExcludeServerUserRoles == nil {
			testCase.output.ExcludeServerUserRoles = make(map[ServerUserRole]any, 0)
		}

		if testCase.output.CourseUserReferences == nil {
			testCase.output.CourseUserReferences = make(map[string]*ParsedCourseUserReference, 0)
		}

		for _, courseReference := range testCase.output.CourseUserReferences {
			if courseReference.Emails == nil {
				courseReference.Emails = make(map[string]any, 0)
			}

			if courseReference.ExcludeEmails == nil {
				courseReference.ExcludeEmails = make(map[string]any, 0)
			}

			if courseReference.CourseUserRoles == nil {
				courseReference.CourseUserRoles = make(map[CourseUserRole]any, 0)
			}

			if courseReference.ExcludeCourseUserRoles == nil {
				courseReference.ExcludeCourseUserRoles = make(map[CourseUserRole]any, 0)
			}
		}

		if !reflect.DeepEqual(testCase.output, result) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.output), util.MustToJSONIndent(result))
			continue
		}
	}
}

func TestResolveServerUserEmails(test *testing.T) {
	extraTestCases := []resolveServerUserTestCase{
		// Outside Emails
		{
			&ParsedServerUserReference{
				Emails: map[string]any{
					"outside-email@test.edulinq.org": nil,
				},
			},
			defaultServerUsers,
			[]string{"outside-email@test.edulinq.org"},
		},
		{
			&ParsedServerUserReference{
				Emails: map[string]any{
					"outside-email@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"outside-email@test.edulinq.org": nil,
				},
			},
			defaultServerUsers,
			[]string{},
		},
	}

	testCases := make([]resolveServerUserTestCase, len(resolveServerUserTestCases))
	copy(resolveServerUserTestCases, testCases)

	testCases = append(testCases, extraTestCases...)

	for i, testCase := range testCases {
		actualOutput := ResolveServerUserEmails(testCase.users, testCase.reference)

		if !reflect.DeepEqual(testCase.expectedOutput, actualOutput) {
			test.Errorf("Case %d: Incorrect Output. Expected: '%v', Actual: '%v'.",
				i, util.MustToJSONIndent(testCase.expectedOutput), util.MustToJSONIndent(actualOutput))
			continue
		}
	}
}

func TestResolveServerUsers(test *testing.T) {
	for i, testCase := range resolveServerUserTestCases {
		actualOutput := ResolveServerUsers(testCase.users, testCase.reference)

		expectedOutput := []*ServerUser{}

		for _, email := range testCase.expectedOutput {
			user, ok := testCase.users[email]
			if !ok {
				test.Errorf("Case %d: Failed to get expected server user: '%s'.", i, email)
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

func getAllServerRoles() map[ServerUserRole]any {
	commonServerRoles := GetCommonServerUserRolesCopy()

	allServerRoles := make(map[ServerUserRole]any, len(commonServerRoles))
	for _, role := range commonServerRoles {
		allServerRoles[role] = nil
	}

	return allServerRoles
}
