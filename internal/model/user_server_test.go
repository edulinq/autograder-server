package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestUserServerUserValidate(test *testing.T) {
	testCases := []struct {
		Email    string
		Name     *string
		Role     ServerUserRole
		Salt     *string
		Tokens   []string
		Roles    map[string]CourseUserRole
		LMSIDs   map[string]string
		Expected *ServerUser
	}{
		// Base
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},

		// Email
		{
			" " + baseTestServerUser.Email + " ",
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			"",
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			nil,
		},

		// Name
		{
			baseTestServerUser.Email,
			util.StringPointer(" " + *baseTestServerUser.Name + " "),
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			nil,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserName(baseTestServerUser, nil),
		},
		{
			baseTestServerUser.Email,
			util.StringPointer(""),
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			nil,
		},

		// Role
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			ServerRoleUnknown,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserRole(baseTestServerUser, ServerRoleUnknown),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			ServerRoleRoot,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			nil,
		},

		// Salt
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			util.StringPointer(" " + *baseTestServerUser.Salt + " "),
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			util.StringPointer(strings.ToUpper(*baseTestServerUser.Salt)),
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			nil,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserSalt(baseTestServerUser, nil),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			util.StringPointer(""),
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserSalt(baseTestServerUser, util.StringPointer("")),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			util.StringPointer("nothex"),
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			nil,
		},

		// Tokens
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			[]string{" " + baseTestServerUser.Tokens[0] + " "},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			[]string{strings.ToUpper(baseTestServerUser.Tokens[0])},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			[]string{baseTestServerUser.Tokens[0], baseTestServerUser.Tokens[0]},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			[]string{"ffff", "0000"},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserTokens(baseTestServerUser, []string{"0000", "ffff"}),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			nil,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserTokens(baseTestServerUser, []string{}),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			[]string{""},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			nil,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			[]string{"ZZ"},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			nil,
		},

		// Course Roles
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			nil,
			baseTestServerUser.LMSIDs,
			setServerCourseUserRoles(baseTestServerUser, make(map[string]CourseUserRole, 0)),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			map[string]CourseUserRole{" course101 ": RoleStudent},
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			map[string]CourseUserRole{"COURSE101": RoleStudent},
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			map[string]CourseUserRole{"": RoleStudent},
			baseTestServerUser.LMSIDs,
			nil,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			map[string]CourseUserRole{"course101": RoleUnknown},
			baseTestServerUser.LMSIDs,
			nil,
		},

		// LMS IDs
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			nil,
			setServerUserLMSIDs(baseTestServerUser, make(map[string]string, 0)),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			map[string]string{" course101 ": "alice"},
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			map[string]string{"course101": " alice "},
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			map[string]string{"COURSE101": "alice"},
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			map[string]string{"": "alice"},
			nil,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			map[string]string{"course101": ""},
			nil,
		},
	}

	for i, testCase := range testCases {
		user := &ServerUser{
			Email:  testCase.Email,
			Name:   testCase.Name,
			Role:   testCase.Role,
			Salt:   testCase.Salt,
			Tokens: testCase.Tokens,
			Roles:  testCase.Roles,
			LMSIDs: testCase.LMSIDs,
		}

		err := user.Validate()
		if err != nil {
			if testCase.Expected == nil {
				// Expected failure.
				continue
			}

			test.Errorf("Case %d: User did not validate: '%v'.", i, err)
			continue
		}

		if testCase.Expected == nil {
			test.Errorf("Case %d: Expected failure did not happen: '%s'.", i, util.MustToJSONIndent(user))
			continue
		}

		if !reflect.DeepEqual(testCase.Expected, user) {
			test.Errorf("Case %d: User not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.Expected), util.MustToJSONIndent(user))
			continue
		}
	}
}

func TestUserServerUserName(test *testing.T) {
	testCases := []struct {
		BaseName     *string
		ResultName   string
		FallbackName string
	}{
		{util.StringPointer("alice"), "alice", "alice"},
		{util.StringPointer(""), "", "alice@test.com"},
		{nil, "", "alice@test.com"},
	}

	for i, testCase := range testCases {
		user := setServerUserName(baseTestServerUser, testCase.BaseName)

		// Skip validation.

		resultName := user.GetName(false)
		fallbackName := user.GetName(true)
		displayName := user.GetDisplayName()

		if testCase.ResultName != resultName {
			test.Errorf("Case %d: Result name not as expected. Expected: '%s', Actual: '%s'.", i, testCase.ResultName, resultName)
			continue
		}

		if testCase.FallbackName != fallbackName {
			test.Errorf("Case %d: Fallback name not as expected. Expected: '%s', Actual: '%s'.", i, testCase.FallbackName, fallbackName)
			continue
		}

		if testCase.FallbackName != displayName {
			test.Errorf("Case %d: Display name not as expected. Expected: '%s', Actual: '%s'.", i, testCase.FallbackName, displayName)
			continue
		}
	}
}

func TestUserServerUserGetCourseUser(test *testing.T) {
	testCases := []struct {
		ServerUser *ServerUser
		CourseUser *CourseUser
		CourseID   string
		HasError   bool
	}{
		// Base
		{
			baseTestServerUser,
			baseTestCourseUser,
			"course101",
			false,
		},

		// Not Enrolled
		{
			baseTestServerUser,
			nil,
			"ZZZ",
			false,
		},

		// No LMS ID
		{
			setServerUserLMSIDs(baseTestServerUser, make(map[string]string, 0)),
			setCourseUserLMSID(baseTestCourseUser, nil),
			"course101",
			false,
		},

		// Validation Error
		{
			setServerCourseUserRoles(baseTestServerUser, map[string]CourseUserRole{"course101": RoleUnknown}),
			nil,
			"course101",
			true,
		},
	}

	for i, testCase := range testCases {
		courseUser, err := testCase.ServerUser.GetCourseUser(testCase.CourseID)
		if err != nil {
			if !testCase.HasError {
				test.Errorf("Case %d: Failed to get course user: '%v'.", i, err)
			}

			continue
		}

		if !reflect.DeepEqual(testCase.CourseUser, courseUser) {
			test.Errorf("Case %d: Course user not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.CourseUser), util.MustToJSONIndent(courseUser))
			continue
		}
	}
}

func TestUserServerUserMerge(test *testing.T) {
	testCases := []struct {
		Source   *ServerUser
		Other    *ServerUser
		Expected *ServerUser
	}{
		// No Change
		{
			baseTestServerUser,
			baseTestServerUser,
			baseTestServerUser,
		},

		// Noop
		{
			baseTestServerUser,
			nil,
			nil,
		},
		{
			baseTestServerUser,
			setServerUserEmail(baseTestServerUser, "zzz"),
			nil,
		},

		// Validation Error
		{
			baseTestServerUser,
			setServerUserSalt(baseTestServerUser, util.StringPointer("ZZZ")),
			nil,
		},

		// Name
		{
			baseTestServerUser,
			setServerUserName(baseTestServerUser, util.StringPointer("foo")),
			setServerUserName(baseTestServerUser, util.StringPointer("foo")),
		},
		{
			baseTestServerUser,
			setServerUserName(baseTestServerUser, nil),
			baseTestServerUser,
		},
		{
			baseTestServerUser,
			setServerUserName(baseTestServerUser, util.StringPointer("")),
			nil,
		},

		// Role
		{
			baseTestServerUser,
			setServerUserRole(baseTestServerUser, ServerRoleUnknown),
			baseTestServerUser,
		},
		{
			baseTestServerUser,
			setServerUserRole(baseTestServerUser, ServerRoleOwner),
			setServerUserRole(baseTestServerUser, ServerRoleOwner),
		},
		{
			baseTestServerUser,
			setServerUserRole(baseTestServerUser, ServerRoleRoot),
			nil,
		},

		// Salt
		{
			baseTestServerUser,
			setServerUserSalt(baseTestServerUser, util.StringPointer("1234")),
			setServerUserSalt(baseTestServerUser, util.StringPointer("1234")),
		},
		{
			baseTestServerUser,
			setServerUserSalt(baseTestServerUser, nil),
			baseTestServerUser,
		},

		// Tokens
		{
			// No tokens will be added.
			baseTestServerUser,
			setServerUserTokens(baseTestServerUser, []string{}),
			baseTestServerUser,
		},
		{
			// No tokens will be added.
			baseTestServerUser,
			setServerUserTokens(baseTestServerUser, nil),
			baseTestServerUser,
		},
		{
			// The token will be added.
			baseTestServerUser,
			setServerUserTokens(baseTestServerUser, []string{"1234"}),
			setServerUserTokens(baseTestServerUser, []string{"1234", "abc123"}),
		},
		{
			// Token is duplicate, will be skipped.
			baseTestServerUser,
			setServerUserTokens(baseTestServerUser, []string{"abc123"}),
			baseTestServerUser,
		},

		// Course Roles
		{
			// No roles will be added.
			baseTestServerUser,
			setServerCourseUserRoles(baseTestServerUser, map[string]CourseUserRole{}),
			baseTestServerUser,
		},
		{
			// No roles will be added.
			baseTestServerUser,
			setServerCourseUserRoles(baseTestServerUser, nil),
			baseTestServerUser,
		},
		{
			// Role will be added.
			baseTestServerUser,
			setServerCourseUserRoles(baseTestServerUser, map[string]CourseUserRole{"foo": RoleStudent}),
			setServerCourseUserRoles(baseTestServerUser, map[string]CourseUserRole{"course101": RoleStudent, "foo": RoleStudent}),
		},
		{
			// Existing role will be overwritten.
			baseTestServerUser,
			setServerCourseUserRoles(baseTestServerUser, map[string]CourseUserRole{"course101": RoleGrader}),
			setServerCourseUserRoles(baseTestServerUser, map[string]CourseUserRole{"course101": RoleGrader}),
		},

		// LMSIDs
		{
			// No lms ids will be added.
			baseTestServerUser,
			setServerUserLMSIDs(baseTestServerUser, map[string]string{}),
			baseTestServerUser,
		},
		{
			// No lms ids will be added.
			baseTestServerUser,
			setServerUserLMSIDs(baseTestServerUser, nil),
			baseTestServerUser,
		},
		{
			// LMSID will be added.
			baseTestServerUser,
			setServerUserLMSIDs(baseTestServerUser, map[string]string{"foo": "bar"}),
			setServerUserLMSIDs(baseTestServerUser, map[string]string{"course101": "alice", "foo": "bar"}),
		},
		{
			// Existing lms ids will be overwritten.
			baseTestServerUser,
			setServerUserLMSIDs(baseTestServerUser, map[string]string{"course101": "bar"}),
			setServerUserLMSIDs(baseTestServerUser, map[string]string{"course101": "bar"}),
		},
	}

	for i, testCase := range testCases {
		source := testCase.Source.Clone()

		err := source.Merge(testCase.Other)
		if err != nil {
			if testCase.Expected != nil {
				test.Errorf("Case %d: Failed to merge user: '%v'.", i, err)
			}

			continue
		}

		if !reflect.DeepEqual(testCase.Expected, source) {
			test.Errorf("Case %d: Merged user not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.Expected), util.MustToJSONIndent(source))
			continue
		}
	}
}

func TestUserServerUserMustToRow(test *testing.T) {
	testCases := []struct {
		User     *ServerUser
		Expected []string
	}{
		// Base
		{
			baseTestServerUser,
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},

		// Email
		{
			setServerUserEmail(baseTestServerUser, "foo"),
			[]string{
				"foo",
				"Alice",
				"user",
				"abc123",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},
		{
			setServerUserEmail(baseTestServerUser, ""),
			[]string{
				"",
				"Alice",
				"user",
				"abc123",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},

		// Name
		{
			setServerUserName(baseTestServerUser, util.StringPointer("foo")),
			[]string{
				"alice@test.com",
				"foo",
				"user",
				"abc123",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},
		{
			setServerUserName(baseTestServerUser, nil),
			[]string{
				"alice@test.com",
				"",
				"user",
				"abc123",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},

		// Role
		{
			setServerUserRole(baseTestServerUser, ServerRoleOwner),
			[]string{
				"alice@test.com",
				"Alice",
				"owner",
				"abc123",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},

		// Salt
		{
			setServerUserSalt(baseTestServerUser, util.StringPointer("aaaa")),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"aaaa",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},
		{
			setServerUserSalt(baseTestServerUser, nil),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},

		// Tokens
		{
			setServerUserTokens(baseTestServerUser, []string{"aaaa"}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				`["aaaa"]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},
		{
			setServerUserTokens(baseTestServerUser, []string{}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				`[]`,
				`{"course101":"student"}`,
				`{"course101":"alice"}`,
			},
		},

		// Course Roles
		{
			setServerCourseUserRoles(baseTestServerUser, map[string]CourseUserRole{"foo": RoleGrader}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				`["abc123"]`,
				`{"foo":"grader"}`,
				`{"course101":"alice"}`,
			},
		},
		{
			setServerCourseUserRoles(baseTestServerUser, map[string]CourseUserRole{}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				`["abc123"]`,
				`{}`,
				`{"course101":"alice"}`,
			},
		},

		// LMS IDs
		{
			setServerUserLMSIDs(baseTestServerUser, map[string]string{"foo": "bar"}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{"foo":"bar"}`,
			},
		},
		{
			setServerUserLMSIDs(baseTestServerUser, map[string]string{}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				`["abc123"]`,
				`{"course101":"student"}`,
				`{}`,
			},
		},
	}

	for i, testCase := range testCases {
		actual := testCase.User.MustToRow()

		if !reflect.DeepEqual(testCase.Expected, actual) {
			test.Errorf("Case %d: Row not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.Expected), util.MustToJSONIndent(actual))
			continue
		}
	}
}

func setServerUserEmail(user *ServerUser, email string) *ServerUser {
	newUser := *user
	newUser.Email = email
	return &newUser
}

func setServerUserName(user *ServerUser, name *string) *ServerUser {
	newUser := *user
	newUser.Name = name
	return &newUser
}

func setServerUserRole(user *ServerUser, role ServerUserRole) *ServerUser {
	newUser := *user
	newUser.Role = role
	return &newUser
}

func setServerUserSalt(user *ServerUser, salt *string) *ServerUser {
	newUser := *user
	newUser.Salt = salt
	return &newUser
}

func setServerUserTokens(user *ServerUser, tokens []string) *ServerUser {
	newUser := *user
	newUser.Tokens = tokens
	return &newUser
}

func setServerCourseUserRoles(user *ServerUser, roles map[string]CourseUserRole) *ServerUser {
	newUser := *user
	newUser.Roles = roles
	return &newUser
}

func setServerUserLMSIDs(user *ServerUser, lmsIDs map[string]string) *ServerUser {
	newUser := *user
	newUser.LMSIDs = lmsIDs
	return &newUser
}

var baseTestServerUser *ServerUser = &ServerUser{
	Email:  "alice@test.com",
	Name:   util.StringPointer("Alice"),
	Role:   ServerRoleUser,
	Salt:   util.StringPointer("abc123"),
	Tokens: []string{"abc123"},
	Roles:  map[string]CourseUserRole{"course101": RoleStudent},
	LMSIDs: map[string]string{"course101": "alice"},
}

var minimalTestServerUser *ServerUser = setServerUserRole(setServerUserTokens(setServerUserSalt(baseTestServerUser, nil), make([]string, 0)), ServerRoleUnknown)
