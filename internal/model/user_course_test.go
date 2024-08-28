package model

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestUserCourseUserValidate(test *testing.T) {
	testCases := []struct {
		Email    string
		Name     *string
		Role     CourseUserRole
		LMSID    *string
		Expected *CourseUser
	}{
		// Base
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			baseTestCourseUser,
		},

		// Email
		{
			" " + baseTestCourseUser.Email + " ",
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			baseTestCourseUser,
		},
		{
			"",
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			nil,
		},

		// Name
		{
			baseTestCourseUser.Email,
			util.StringPointer(" " + *baseTestCourseUser.Name + " "),
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			baseTestCourseUser,
		},
		{
			baseTestCourseUser.Email,
			nil,
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			setCourseUserName(baseTestCourseUser, nil),
		},
		{
			baseTestCourseUser.Email,
			util.StringPointer(""),
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			setCourseUserName(baseTestCourseUser, util.StringPointer("")),
		},

		// Role
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			CourseRoleUnknown,
			baseTestCourseUser.LMSID,
			nil,
		},

		// LMS ID
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			util.StringPointer(" " + *baseTestCourseUser.LMSID + " "),
			baseTestCourseUser,
		},
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			nil,
			setCourseUserLMSID(baseTestCourseUser, nil),
		},
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			util.StringPointer(""),
			setCourseUserLMSID(baseTestCourseUser, util.StringPointer("")),
		},
	}

	for i, testCase := range testCases {
		user := &CourseUser{
			Email: testCase.Email,
			Name:  testCase.Name,
			Role:  testCase.Role,
			LMSID: testCase.LMSID,
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

func TestUserCourseUserName(test *testing.T) {
	testCases := []struct {
		BaseName     *string
		ResultName   string
		FallbackName string
	}{
		{util.StringPointer("alice"), "alice", "alice"},
		{util.StringPointer(""), "", "alice@test.edulinq.org"},
		{nil, "", "alice@test.edulinq.org"},
	}

	for i, testCase := range testCases {
		user := setCourseUserName(baseTestCourseUser, testCase.BaseName)
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

func TestUserCourseUserToServerUser(test *testing.T) {
	minConversionCourseUser := &ServerUser{
		Email:  "alice@test.edulinq.org",
		Name:   util.StringPointer("Alice"),
		Role:   ServerRoleUnknown,
		Salt:   nil,
		Tokens: []*Token{},
		CourseInfo: map[string]*UserCourseInfo{
			"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: util.StringPointer("alice")},
		},
	}

	testCases := []struct {
		CourseUser *CourseUser
		ServerUser *ServerUser
		CourseID   string
		HasError   bool
	}{
		// Base
		{
			baseTestCourseUser,
			minConversionCourseUser,
			"course101",
			false,
		},

		// No LMS ID
		{
			setCourseUserLMSID(baseTestCourseUser, nil),
			setServerUserCourseInfo(minConversionCourseUser, map[string]*UserCourseInfo{
				"course101": &UserCourseInfo{Role: CourseRoleStudent, LMSID: nil},
			}),
			"course101",
			false,
		},

		// Validation Error
		{
			setCourseCourseUserRole(baseTestCourseUser, CourseRoleUnknown),
			nil,
			"course101",
			true,
		},
	}

	for i, testCase := range testCases {
		serverUser, err := testCase.CourseUser.ToServerUser(testCase.CourseID)
		if err != nil {
			if !testCase.HasError {
				test.Errorf("Case %d: Failed to get server user: '%v'.", i, err)
			}

			continue
		}

		if !reflect.DeepEqual(testCase.ServerUser, serverUser) {
			test.Errorf("Case %d: Server user not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.ServerUser), util.MustToJSONIndent(serverUser))
			continue
		}
	}
}

func TestUserCourseUserMustToRow(test *testing.T) {
	testCases := []struct {
		User     *CourseUser
		Expected []string
	}{
		// Base
		{
			baseTestCourseUser,
			[]string{
				"alice@test.edulinq.org",
				"Alice",
				"student",
				"alice",
			},
		},

		// Email
		{
			setCourseUserEmail(baseTestCourseUser, "foo"),
			[]string{
				"foo",
				"Alice",
				"student",
				"alice",
			},
		},
		{
			setCourseUserEmail(baseTestCourseUser, ""),
			[]string{
				"",
				"Alice",
				"student",
				"alice",
			},
		},

		// Name
		{
			setCourseUserName(baseTestCourseUser, util.StringPointer("foo")),
			[]string{
				"alice@test.edulinq.org",
				"foo",
				"student",
				"alice",
			},
		},
		{
			setCourseUserName(baseTestCourseUser, nil),
			[]string{
				"alice@test.edulinq.org",
				"",
				"student",
				"alice",
			},
		},

		// Role
		{
			setCourseCourseUserRole(baseTestCourseUser, CourseRoleGrader),
			[]string{
				"alice@test.edulinq.org",
				"Alice",
				"grader",
				"alice",
			},
		},

		// LMS ID
		{
			setCourseUserLMSID(baseTestCourseUser, util.StringPointer("foo")),
			[]string{
				"alice@test.edulinq.org",
				"Alice",
				"student",
				"foo",
			},
		},
		{
			setCourseUserLMSID(baseTestCourseUser, nil),
			[]string{
				"alice@test.edulinq.org",
				"Alice",
				"student",
				"",
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

func setCourseUserEmail(user *CourseUser, email string) *CourseUser {
	newUser := *user
	newUser.Email = email
	return &newUser
}

func setCourseUserName(user *CourseUser, name *string) *CourseUser {
	newUser := *user
	newUser.Name = name
	return &newUser
}

func setCourseCourseUserRole(user *CourseUser, role CourseUserRole) *CourseUser {
	newUser := *user
	newUser.Role = role
	return &newUser
}

func setCourseUserLMSID(user *CourseUser, lmsID *string) *CourseUser {
	newUser := *user
	newUser.LMSID = lmsID
	return &newUser
}

var baseTestCourseUser *CourseUser = &CourseUser{
	Email: "alice@test.edulinq.org",
	Name:  util.StringPointer("Alice"),
	Role:  CourseRoleStudent,
	LMSID: util.StringPointer("alice"),
}
