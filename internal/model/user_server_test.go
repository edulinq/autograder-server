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
		Salt     *string
		Tokens   []string
		Roles    map[string]UserRole
		LMSIDs   map[string]string
		Expected *ServerUser
	}{
		// Base
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
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
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			"",
			baseTestServerUser.Name,
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
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			nil,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserName(baseTestServerUser, nil),
		},
		{
			baseTestServerUser.Email,
			util.StringPointer(""),
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserName(baseTestServerUser, util.StringPointer("")),
		},

		// Salt
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			util.StringPointer(" " + *baseTestServerUser.Salt + " "),
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			util.StringPointer(strings.ToUpper(*baseTestServerUser.Salt)),
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			nil,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserSalt(baseTestServerUser, nil),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			util.StringPointer(""),
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserSalt(baseTestServerUser, util.StringPointer("")),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
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
			baseTestServerUser.Salt,
			[]string{" " + baseTestServerUser.Tokens[0] + " "},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			[]string{strings.ToUpper(baseTestServerUser.Tokens[0])},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			[]string{baseTestServerUser.Tokens[0], baseTestServerUser.Tokens[0]},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			[]string{"ffff", "0000"},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserTokens(baseTestServerUser, []string{"0000", "ffff"}),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			nil,
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			setServerUserTokens(baseTestServerUser, []string{}),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			[]string{""},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			nil,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			[]string{"ZZ"},
			baseTestServerUser.Roles,
			baseTestServerUser.LMSIDs,
			nil,
		},

		// Roles
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			nil,
			baseTestServerUser.LMSIDs,
			setServerUserRoles(baseTestServerUser, make(map[string]UserRole, 0)),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			map[string]UserRole{" course101 ": RoleStudent},
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			map[string]UserRole{"COURSE101": RoleStudent},
			baseTestServerUser.LMSIDs,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			map[string]UserRole{"": RoleStudent},
			baseTestServerUser.LMSIDs,
			nil,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			map[string]UserRole{"course101": RoleUnknown},
			baseTestServerUser.LMSIDs,
			nil,
		},

		// LMS IDs
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			nil,
			setServerUserLMSIDs(baseTestServerUser, make(map[string]string, 0)),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			map[string]string{" course101 ": "alice"},
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			map[string]string{"course101": " alice "},
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			map[string]string{"COURSE101": "alice"},
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Salt,
			baseTestServerUser.Tokens,
			baseTestServerUser.Roles,
			map[string]string{"": "alice"},
			nil,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
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
			test.Errorf("User not as expected. Expected: '%s', Actual: '%s'.",
				util.MustToJSONIndent(testCase.Expected), util.MustToJSONIndent(user))
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
		err := user.Validate()
		if err != nil {
			test.Errorf("Case %d: Failed to validate user: '%v'.", i, err)
			continue
		}

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
			setServerUserRoles(baseTestServerUser, map[string]UserRole{"course101": RoleUnknown}),
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
			test.Errorf("Course user not as expected. Expected: '%s', Actual: '%s'.",
				util.MustToJSONIndent(testCase.CourseUser), util.MustToJSONIndent(courseUser))
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

		// Roles
		{
			// No roles will be added.
			baseTestServerUser,
			setServerUserRoles(baseTestServerUser, map[string]UserRole{}),
			baseTestServerUser,
		},
		{
			// No roles will be added.
			baseTestServerUser,
			setServerUserRoles(baseTestServerUser, nil),
			baseTestServerUser,
		},
		{
			// Role will be added.
			baseTestServerUser,
			setServerUserRoles(baseTestServerUser, map[string]UserRole{"foo": RoleStudent}),
			setServerUserRoles(baseTestServerUser, map[string]UserRole{"course101": RoleStudent, "foo": RoleStudent}),
		},
		{
			// Existing role will be overwritten.
			baseTestServerUser,
			setServerUserRoles(baseTestServerUser, map[string]UserRole{"course101": RoleGrader}),
			setServerUserRoles(baseTestServerUser, map[string]UserRole{"course101": RoleGrader}),
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
			test.Errorf("Merged user not as expected. Expected: '%s', Actual: '%s'.",
				util.MustToJSONIndent(testCase.Expected), util.MustToJSONIndent(source))
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

func setServerUserRoles(user *ServerUser, roles map[string]UserRole) *ServerUser {
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
	Salt:   util.StringPointer("abc123"),
	Tokens: []string{"abc123"},
	Roles:  map[string]UserRole{"course101": RoleStudent},
	LMSIDs: map[string]string{"course101": "alice"},
}

var minimalTestServerUser *ServerUser = setServerUserTokens(setServerUserSalt(baseTestServerUser, nil), make([]string, 0))
