package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestServerUserValidate(test *testing.T) {
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
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			baseTestUser,
		},

		// Email
		{
			" " + baseTestUser.Email + " ",
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			baseTestUser,
		},
		{
			"",
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			nil,
		},

		// Name
		{
			baseTestUser.Email,
			stringPointer(" " + *baseTestUser.Name + " "),
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			baseTestUser,
		},
		{
			baseTestUser.Email,
			nil,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			setName(baseTestUser, nil),
		},
		{
			baseTestUser.Email,
			stringPointer(""),
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			setName(baseTestUser, stringPointer("")),
		},

		// Salt
		{
			baseTestUser.Email,
			baseTestUser.Name,
			stringPointer(" " + *baseTestUser.Salt + " "),
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			stringPointer(strings.ToUpper(*baseTestUser.Salt)),
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			nil,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			setSalt(baseTestUser, nil),
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			stringPointer(""),
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			setSalt(baseTestUser, stringPointer("")),
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			stringPointer("nothex"),
			baseTestUser.Tokens,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			nil,
		},

		// Tokens
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			[]string{" " + baseTestUser.Tokens[0] + " "},
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			[]string{strings.ToUpper(baseTestUser.Tokens[0])},
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			[]string{baseTestUser.Tokens[0], baseTestUser.Tokens[0]},
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			[]string{"ffff", "0000"},
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			setTokens(baseTestUser, []string{"0000", "ffff"}),
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			nil,
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			setTokens(baseTestUser, []string{}),
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			[]string{""},
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			nil,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			[]string{"ZZ"},
			baseTestUser.Roles,
			baseTestUser.LMSIDs,
			nil,
		},

		// Roles
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			nil,
			baseTestUser.LMSIDs,
			setRoles(baseTestUser, make(map[string]UserRole, 0)),
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			map[string]UserRole{" foo ": RoleStudent},
			baseTestUser.LMSIDs,
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			map[string]UserRole{"FOO": RoleStudent},
			baseTestUser.LMSIDs,
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			map[string]UserRole{"": RoleStudent},
			baseTestUser.LMSIDs,
			nil,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			map[string]UserRole{"foo": RoleUnknown},
			baseTestUser.LMSIDs,
			nil,
		},

		// LMS IDs
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			nil,
			setLMSIDs(baseTestUser, make(map[string]string, 0)),
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			map[string]string{" foo ": "alice"},
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			map[string]string{"foo": " alice "},
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			map[string]string{"FOO": "alice"},
			baseTestUser,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			map[string]string{"": "alice"},
			nil,
		},
		{
			baseTestUser.Email,
			baseTestUser.Name,
			baseTestUser.Salt,
			baseTestUser.Tokens,
			baseTestUser.Roles,
			map[string]string{"foo": ""},
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

func stringPointer(text string) *string {
	return &text
}

func setName(user *ServerUser, name *string) *ServerUser {
	newUser := *user
	newUser.Name = name
	return &newUser
}

func setSalt(user *ServerUser, salt *string) *ServerUser {
	newUser := *user
	newUser.Salt = salt
	return &newUser
}

func setTokens(user *ServerUser, tokens []string) *ServerUser {
	newUser := *user
	newUser.Tokens = tokens
	return &newUser
}

func setRoles(user *ServerUser, roles map[string]UserRole) *ServerUser {
	newUser := *user
	newUser.Roles = roles
	return &newUser
}

func setLMSIDs(user *ServerUser, lmsIDs map[string]string) *ServerUser {
	newUser := *user
	newUser.LMSIDs = lmsIDs
	return &newUser
}

var baseTestUser *ServerUser = &ServerUser{
	Email:  "alice@test.com",
	Name:   stringPointer("Alice"),
	Salt:   stringPointer("abc123"),
	Tokens: []string{"abc123"},
	Roles:  map[string]UserRole{"foo": RoleStudent},
	LMSIDs: map[string]string{"foo": "alice"},
}
