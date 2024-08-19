package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

const BASE_SALT string = "abc123"

var baseTestToken *Token = MustNewToken("321cba", BASE_SALT, TokenSourceServer, "test token")
var baseTestPasswordToken *Token = MustNewToken("alice", BASE_SALT, TokenSourcePassword, "password")

func TestUserServerUserValidate(test *testing.T) {
	testCases := []struct {
		Email      string
		Name       *string
		Role       ServerUserRole
		Salt       *string
		Password   *Token
		Tokens     []*Token
		CourseInfo map[string]*UserCourseInfo
		Expected   *ServerUser
	}{
		// Base
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			baseTestServerUser,
		},

		// Email
		{
			" " + baseTestServerUser.Email + " ",
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			baseTestServerUser,
		},
		{
			"",
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			nil,
		},

		// Name
		{
			baseTestServerUser.Email,
			util.StringPointer(" " + *baseTestServerUser.Name + " "),
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			nil,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			setServerUserName(baseTestServerUser, nil),
		},
		{
			baseTestServerUser.Email,
			util.StringPointer(""),
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			nil,
		},

		// Server Role
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			ServerRoleUnknown,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			setServerUserRole(baseTestServerUser, ServerRoleUnknown),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			ServerRoleRoot,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			nil,
		},

		// Salt
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			util.StringPointer(" " + *baseTestServerUser.Salt + " "),
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			util.StringPointer(strings.ToUpper(*baseTestServerUser.Salt)),
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			nil,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			setServerUserSalt(baseTestServerUser, nil),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			util.StringPointer(""),
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			setServerUserSalt(baseTestServerUser, util.StringPointer("")),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			util.StringPointer("nothex"),
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			nil,
		},

		// Password

		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			nil,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			setServerUserPassword(baseTestServerUser, nil),
		},

		// Non-password token used as password.
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestToken,
			baseTestServerUser.Tokens,
			baseTestServerUser.CourseInfo,
			nil,
		},

		// Tokens

		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			nil,
			baseTestServerUser.CourseInfo,
			setServerUserTokens(baseTestServerUser, []*Token{}),
		},

		// Bad hash (not hex).
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			[]*Token{&Token{
				HexDigest: "ZZZ",
				Source:    TokenSourceServer,
			}},
			baseTestServerUser.CourseInfo,
			nil,
		},

		// Nil token.
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			[]*Token{nil},
			baseTestServerUser.CourseInfo,
			nil,
		},

		// Token is password.
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			[]*Token{baseTestPasswordToken},
			baseTestServerUser.CourseInfo,
			nil,
		},

		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			[]*Token{baseTestToken, baseTestToken},
			baseTestServerUser.CourseInfo,
			baseTestServerUser,
		},

		// Course Info
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			nil,
			setServerUserCourseInfo(baseTestServerUser, make(map[string]*UserCourseInfo, 0)),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			map[string]*UserCourseInfo{" course101 ": baseUserCourseInfo},
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			map[string]*UserCourseInfo{"COURSE101": baseUserCourseInfo},
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			map[string]*UserCourseInfo{"": baseUserCourseInfo},
			nil,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: nil},
			},
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: nil},
			}),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: util.StringPointer(" alice ")},
			},
			baseTestServerUser,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleUnknown, LMSID: nil},
			},
			nil,
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: util.StringPointer("")},
			},
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: nil},
			}),
		},
		{
			baseTestServerUser.Email,
			baseTestServerUser.Name,
			baseTestServerUser.Role,
			baseTestServerUser.Salt,
			baseTestServerUser.Password,
			baseTestServerUser.Tokens,
			map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: util.StringPointer(" ")},
			},
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: nil},
			}),
		},
	}

	for i, testCase := range testCases {
		user := &ServerUser{
			Email:      testCase.Email,
			Name:       testCase.Name,
			Role:       testCase.Role,
			Salt:       testCase.Salt,
			Password:   testCase.Password,
			Tokens:     testCase.Tokens,
			CourseInfo: testCase.CourseInfo,
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

func TestUserServerUserToCourseUser(test *testing.T) {
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
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{"course101": &UserCourseInfo{Role: CourseRoleStudent}}),
			setCourseUserLMSID(baseTestCourseUser, nil),
			"course101",
			false,
		},

		// Validation Error
		{
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{"course101": &UserCourseInfo{Role: CourseRoleUnknown}}),
			nil,
			"course101",
			true,
		},
	}

	for i, testCase := range testCases {
		courseUser, err := testCase.ServerUser.ToCourseUser(testCase.CourseID)
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
	baseToken := baseTestServerUser.Tokens[0]
	newToken := MustNewToken("aa", "abc123", TokenSourceServer, "test token 2")
	newPasswordToken := MustNewToken("alice123", BASE_SALT, TokenSourcePassword, "password")

	testCases := []struct {
		Source   *ServerUser
		Other    *ServerUser
		Changed  bool
		Expected *ServerUser
	}{
		// No Change.
		{
			baseTestServerUser,
			minimalTestServerUser,
			false,
			baseTestServerUser,
		},

		// Noop
		{
			baseTestServerUser,
			nil,
			false,
			nil,
		},
		{
			baseTestServerUser,
			setServerUserEmail(minimalTestServerUser, "zzz"),
			false,
			nil,
		},

		// Validation Error
		{
			baseTestServerUser,
			setServerUserSalt(minimalTestServerUser, util.StringPointer("ZZZ")),
			false,
			nil,
		},

		// Name
		{
			baseTestServerUser,
			setServerUserName(minimalTestServerUser, util.StringPointer("foo")),
			true,
			setServerUserName(baseTestServerUser, util.StringPointer("foo")),
		},
		{
			baseTestServerUser,
			setServerUserName(minimalTestServerUser, nil),
			false,
			baseTestServerUser,
		},
		{
			setServerUserName(baseTestServerUser, nil),
			setServerUserName(minimalTestServerUser, util.StringPointer("foo")),
			true,
			setServerUserName(baseTestServerUser, util.StringPointer("foo")),
		},
		{
			baseTestServerUser,
			setServerUserName(minimalTestServerUser, util.StringPointer("")),
			true,
			nil,
		},

		// Role
		{
			baseTestServerUser,
			setServerUserRole(minimalTestServerUser, ServerRoleUnknown),
			false,
			baseTestServerUser,
		},
		{
			baseTestServerUser,
			setServerUserRole(minimalTestServerUser, ServerRoleOwner),
			true,
			setServerUserRole(baseTestServerUser, ServerRoleOwner),
		},
		{
			baseTestServerUser,
			setServerUserRole(minimalTestServerUser, ServerRoleRoot),
			false,
			nil,
		},

		// Salt
		{
			baseTestServerUser,
			setServerUserSalt(minimalTestServerUser, util.StringPointer("1234")),
			true,
			setServerUserSalt(baseTestServerUser, util.StringPointer("1234")),
		},
		{
			baseTestServerUser,
			setServerUserSalt(minimalTestServerUser, nil),
			false,
			baseTestServerUser,
		},
		{
			setServerUserSalt(baseTestServerUser, nil),
			setServerUserSalt(minimalTestServerUser, util.StringPointer("1234")),
			true,
			setServerUserSalt(baseTestServerUser, util.StringPointer("1234")),
		},

		// Password
		{
			baseTestServerUser,
			setServerUserPassword(minimalTestServerUser, nil),
			false,
			baseTestServerUser,
		},
		{
			baseTestServerUser,
			setServerUserPassword(minimalTestServerUser, newPasswordToken),
			true,
			setServerUserPassword(baseTestServerUser, newPasswordToken),
		},

		// Tokens
		{
			// No tokens will be added.
			baseTestServerUser,
			setServerUserTokens(minimalTestServerUser, []*Token{}),
			false,
			baseTestServerUser,
		},
		{
			// No tokens will be added.
			baseTestServerUser,
			setServerUserTokens(minimalTestServerUser, nil),
			false,
			baseTestServerUser,
		},
		{
			// The token will be added.
			baseTestServerUser,
			setServerUserTokens(minimalTestServerUser, []*Token{newToken}),
			true,
			setServerUserTokens(baseTestServerUser, []*Token{baseToken, newToken}),
		},

		// Course Info
		{
			// Nothing will be added.
			baseTestServerUser,
			setServerUserCourseInfo(minimalTestServerUser, map[string]*UserCourseInfo{}),
			false,
			baseTestServerUser,
		},
		{
			// Nothing will be added.
			baseTestServerUser,
			setServerUserCourseInfo(minimalTestServerUser, nil),
			false,
			baseTestServerUser,
		},
		{
			// Info will be added.
			baseTestServerUser,
			setServerUserCourseInfo(minimalTestServerUser, map[string]*UserCourseInfo{
				"foo": &UserCourseInfo{Role: CourseRoleStudent, LMSID: util.StringPointer("bar")},
			}),
			true,
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: util.StringPointer("alice")},
				"foo":       &UserCourseInfo{Role: CourseRoleStudent, LMSID: util.StringPointer("bar")},
			}),
		},
		{
			// Existing info will be overwritten.
			baseTestServerUser,
			setServerUserCourseInfo(minimalTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleGrader, LMSID: util.StringPointer("foo")},
			}),
			true,
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleGrader, LMSID: util.StringPointer("foo")},
			}),
		},
		{
			// Only role will be overwritten.
			baseTestServerUser,
			setServerUserCourseInfo(minimalTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleGrader},
			}),
			true,
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleGrader, LMSID: util.StringPointer("alice")},
			}),
		},
		{
			// Only LMS ID will be overwritten.
			baseTestServerUser,
			setServerUserCourseInfo(minimalTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{LMSID: util.StringPointer("foo")},
			}),
			true,
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: util.StringPointer("foo")},
			}),
		},
	}

	for i, testCase := range testCases {
		source := testCase.Source.Clone()

		changed, err := source.Merge(testCase.Other)
		if err != nil {
			if testCase.Expected != nil {
				test.Errorf("Case %d: Failed to merge user: '%v'.", i, err)
			}

			continue
		}

		if testCase.Changed != changed {
			test.Errorf("Case %d: Changed indicator not as expected. Expected: '%v', Actual: '%v'.",
				i, testCase.Changed, changed)
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
	newPasswordToken := MustNewToken("alice123", BASE_SALT, TokenSourcePassword, "password")

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
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
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
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
			},
		},
		{
			setServerUserEmail(baseTestServerUser, ""),
			[]string{
				"",
				"Alice",
				"user",
				"abc123",
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
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
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
			},
		},
		{
			setServerUserName(baseTestServerUser, nil),
			[]string{
				"alice@test.com",
				"",
				"user",
				"abc123",
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
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
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
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
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
			},
		},
		{
			setServerUserSalt(baseTestServerUser, nil),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"",
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
			},
		},

		// Password
		{
			setServerUserPassword(baseTestServerUser, nil),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				"<nil>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
			},
		},
		{
			setServerUserPassword(baseTestServerUser, newPasswordToken),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
			},
		},

		// Tokens
		{
			setServerUserTokens(baseTestServerUser, []*Token{MustNewToken("aa", "abc123", TokenSourcePassword, "test token 2")}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				"<exists>",
				`["test token 2 (password)"]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
			},
		},
		{
			setServerUserTokens(baseTestServerUser, []*Token{}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				"<exists>",
				`[]`,
				`{"course101":{"role":"student","lms-id":"alice"}}`,
			},
		},

		// Course Info
		{
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{
				"foo": &UserCourseInfo{Role: CourseRoleGrader, LMSID: util.StringPointer("bar")},
			}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				"<exists>",
				`["test token (server)"]`,
				`{"foo":{"role":"grader","lms-id":"bar"}}`,
			},
		},
		{
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{
				"course101": baseUserCourseInfo,
				"foo":       &UserCourseInfo{Role: CourseRoleGrader, LMSID: util.StringPointer("bar")},
			}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				"<exists>",
				`["test token (server)"]`,
				`{"course101":{"role":"student","lms-id":"alice"},"foo":{"role":"grader","lms-id":"bar"}}`,
			},
		},
		{
			setServerUserCourseInfo(baseTestServerUser, map[string]*UserCourseInfo{}),
			[]string{
				"alice@test.com",
				"Alice",
				"user",
				"abc123",
				"<exists>",
				`["test token (server)"]`,
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

func TestUserServerUserAuth(test *testing.T) {
	// Make a user with a pass and two tokens.
	cleartext1, token1 := MustNewRandomToken(BASE_SALT, TokenSourceServer, "test token")
	cleartext2, token2 := MustNewRandomToken(BASE_SALT, TokenSourceServer, "test token")

	user := setServerUserTokens(baseTestServerUser, []*Token{token1, token2})

	passCleartext, err := user.SetRandomPassword()
	if err != nil {
		test.Fatalf("Failed to set test password: '%v'.", err)
	}

	testCases := []struct {
		pass    string
		success bool
	}{
		// Check password.
		{passCleartext, true},

		// Check first token.
		{cleartext1, true},

		// Check second token.
		{cleartext2, true},

		// Check bad input.
		{cleartext1 + cleartext2, false},
		{passCleartext + " ", false},
		{" " + passCleartext, false},
		{cleartext1 + " ", false},
		{" " + cleartext1, false},
	}

	for i, testCase := range testCases {
		success, err := user.Auth(util.Sha256HexFromString(testCase.pass))
		if err != nil {
			test.Errorf("Case %d: Faled to auth: '%v'.", i, err)
			continue
		}

		if testCase.success != success {
			test.Errorf("Case %d: Result not as expected. Expected: %v, Actual: %v.", i, testCase.success, success)
			continue
		}
	}
}

func TestUserServerSetPassword(test *testing.T) {
	user := baseTestServerUser.Clone()
	cleartext := "test"
	input := util.Sha256HexFromString(cleartext)

	changed, err := user.SetPassword(input)
	if err != nil {
		test.Fatalf("Failed to set password: '%v'.", err)
	}

	if !changed {
		test.Fatalf("Password was not changed.")
	}

	success, err := user.Auth(input)
	if err != nil {
		test.Fatalf("Faled to auth: '%v'.", err)
	}

	if !success {
		test.Fatalf("Password was not accepted.")
	}
}

func TestUserServerSetRandomPassword(test *testing.T) {
	user := baseTestServerUser.Clone()

	cleartext, err := user.SetRandomPassword()
	if err != nil {
		test.Fatalf("Failed to set random password: '%v'.", err)
	}

	input := util.Sha256HexFromString(cleartext)

	success, err := user.Auth(input)
	if err != nil {
		test.Fatalf("Faled to auth: '%v'.", err)
	}

	if !success {
		test.Fatalf("Password was not accepted.")
	}
}

func TestUserServerCreateRandomToken(test *testing.T) {
	user := baseTestServerUser.Clone()

	initialTokenCount := len(user.Tokens)

	_, cleartext, err := user.CreateRandomToken("", TokenSourceServer)
	if err != nil {
		test.Fatalf("Failed to create random token: '%v'.", err)
	}

	newTokenCount := len(user.Tokens)

	if newTokenCount != (initialTokenCount + 1) {
		test.Fatalf("Incorrect token count. Expected: %d, Found: %d.", (initialTokenCount + 1), newTokenCount)
	}

	auth, err := user.Auth(util.Sha256HexFromString(cleartext))
	if err != nil {
		test.Fatalf("Failed to perform authentication: '%v'.", err)
	}

	if !auth {
		test.Fatalf("Failed to auth.")
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

func setServerUserPassword(user *ServerUser, password *Token) *ServerUser {
	newUser := *user
	newUser.Password = password
	return &newUser
}

func setServerUserTokens(user *ServerUser, tokens []*Token) *ServerUser {
	newUser := *user
	newUser.Tokens = tokens
	return &newUser
}

func setServerUserCourseInfo(user *ServerUser, info map[string]*UserCourseInfo) *ServerUser {
	newUser := *user
	newUser.CourseInfo = info
	return &newUser
}

var baseUserCourseInfo *UserCourseInfo = &UserCourseInfo{
	Role:  CourseRoleStudent,
	LMSID: util.StringPointer("alice"),
}

var baseTestServerUser *ServerUser = &ServerUser{
	Email:      "alice@test.com",
	Name:       util.StringPointer("Alice"),
	Role:       ServerRoleUser,
	Salt:       util.StringPointer(BASE_SALT),
	Password:   baseTestPasswordToken,
	Tokens:     []*Token{baseTestToken},
	CourseInfo: map[string]*UserCourseInfo{"course101": baseUserCourseInfo},
}

var minimalTestServerUser *ServerUser = &ServerUser{
	Email:      "alice@test.com",
	Name:       nil,
	Role:       ServerRoleUnknown,
	Salt:       nil,
	Password:   nil,
	Tokens:     []*Token{},
	CourseInfo: map[string]*UserCourseInfo{},
}
