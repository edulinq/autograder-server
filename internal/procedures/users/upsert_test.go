package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

var PLACEHOLDER_PASS string = "<placeholder_pass>"
var PLACEHOLDER_SALT *string = util.StringPointer("abcd")

// TEST - Send emails

func TestUpsertUser(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
		// Do not set DryRun, this will be set automatically.
		options      UpsertUsersOptions
		expected     *model.UserOpResult
		expectedUser *model.ServerUser
	}{
		// New user without course.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.com",
						Name:  "new",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
						Pass:  "1234",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email: "new@test.com",
				Added: true,
			},
			expectedUser: &model.ServerUser{
				Email:  "new@test.com",
				Name:   util.StringPointer("new"),
				Salt:   PLACEHOLDER_SALT,
				Role:   model.ServerRoleUser,
				Tokens: []*model.Token{nil},
			},
		},

		// New user with course.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:       "new@test.com",
						Name:        "new",
						Role:        model.GetServerUserRoleString(model.ServerRoleUser),
						Pass:        "1234",
						Course:      "new-course",
						CourseRole:  model.GetCourseUserRoleString(model.RoleStudent),
						CourseLMSID: "new-lms",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "new@test.com",
				Added:    true,
				Enrolled: []string{"new-course"},
			},
			expectedUser: &model.ServerUser{
				Email:  "new@test.com",
				Name:   util.StringPointer("new"),
				Salt:   PLACEHOLDER_SALT,
				Role:   model.ServerRoleUser,
				Tokens: []*model.Token{nil},
				Roles:  map[string]model.CourseUserRole{"new-course": model.RoleStudent},
				LMSIDs: map[string]string{"new-course": "new-lms"},
			},
		},

		// Update user without course.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "student@test.com",
						Name:  "new",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "student@test.com",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:  "student@test.com",
				Name:   util.StringPointer("new"),
				Salt:   PLACEHOLDER_SALT,
				Role:   model.ServerRoleUser,
				Tokens: []*model.Token{nil},
				Roles: map[string]model.CourseUserRole{
					"course-languages":          model.RoleStudent,
					"course-with-lms":           model.RoleStudent,
					"course-without-source":     model.RoleStudent,
					"course101":                 model.RoleStudent,
					"course101-with-zero-limit": model.RoleStudent,
				},
				LMSIDs: map[string]string{
					"course-languages":      "lms-student@test.com",
					"course-with-lms":       "lms-student@test.com",
					"course-without-source": "lms-student@test.com",
				},
			},
		},

		// Update user with course (enroll).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:       "student@test.com",
						Course:      "new-course",
						CourseRole:  model.GetCourseUserRoleString(model.RoleStudent),
						CourseLMSID: "new-lms",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "student@test.com",
				Modified: true,
				Enrolled: []string{"new-course"},
			},
			expectedUser: &model.ServerUser{
				Email:  "student@test.com",
				Name:   util.StringPointer("student"),
				Salt:   PLACEHOLDER_SALT,
				Role:   model.ServerRoleUser,
				Tokens: []*model.Token{nil},
				Roles: map[string]model.CourseUserRole{
					"course-languages":          model.RoleStudent,
					"course-with-lms":           model.RoleStudent,
					"course-without-source":     model.RoleStudent,
					"course101":                 model.RoleStudent,
					"course101-with-zero-limit": model.RoleStudent,
					"new-course":                model.RoleStudent,
				},
				LMSIDs: map[string]string{
					"course-languages":      "lms-student@test.com",
					"course-with-lms":       "lms-student@test.com",
					"course-without-source": "lms-student@test.com",
					"new-course":            "new-lms",
				},
			},
		},

		// Set a new password.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "student@test.com",
						Pass:  "1234",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "student@test.com",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:  "student@test.com",
				Name:   util.StringPointer("student"),
				Salt:   PLACEHOLDER_SALT,
				Role:   model.ServerRoleUser,
				Tokens: []*model.Token{nil, nil},
				Roles: map[string]model.CourseUserRole{
					"course-languages":          model.RoleStudent,
					"course-with-lms":           model.RoleStudent,
					"course-without-source":     model.RoleStudent,
					"course101":                 model.RoleStudent,
					"course101-with-zero-limit": model.RoleStudent,
				},
				LMSIDs: map[string]string{
					"course-languages":      "lms-student@test.com",
					"course-with-lms":       "lms-student@test.com",
					"course-without-source": "lms-student@test.com",
				},
			},
		},

		// Set a duplicate password.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "student@test.com",
						Pass:  util.Sha256HexFromString("student"),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email: "student@test.com",
			},
			expectedUser: &model.ServerUser{
				Email:  "student@test.com",
				Name:   util.StringPointer("student"),
				Salt:   PLACEHOLDER_SALT,
				Role:   model.ServerRoleUser,
				Tokens: []*model.Token{nil},
				Roles: map[string]model.CourseUserRole{
					"course-languages":          model.RoleStudent,
					"course-with-lms":           model.RoleStudent,
					"course-without-source":     model.RoleStudent,
					"course101":                 model.RoleStudent,
					"course101-with-zero-limit": model.RoleStudent,
				},
				LMSIDs: map[string]string{
					"course-languages":      "lms-student@test.com",
					"course-with-lms":       "lms-student@test.com",
					"course-without-source": "lms-student@test.com",
				},
			},
		},

		// New user without password.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.com",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:             "new@test.com",
				Added:             true,
				CleartextPassword: PLACEHOLDER_PASS,
			},
			expectedUser: &model.ServerUser{
				Email:  "new@test.com",
				Salt:   PLACEHOLDER_SALT,
				Role:   model.ServerRoleUser,
				Tokens: []*model.Token{nil},
			},
		},

		// Add a user from a course (but not server) admin.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.com",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.RoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:             "new@test.com",
				Added:             true,
				CleartextPassword: PLACEHOLDER_PASS,
			},
			expectedUser: &model.ServerUser{
				Email:  "new@test.com",
				Salt:   PLACEHOLDER_SALT,
				Role:   model.ServerRoleUser,
				Tokens: []*model.Token{nil},
			},
		},

		// Permission Errors

		// Add a user wihout a any roles.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.com",
					},
				},
			},
			expected: &model.UserOpResult{
				Email:        "",
				SystemErrors: []string{"No authority/roles were provided when adding a user."},
			},
			expectedUser: nil,
		},

		// Add a user with only a course role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.com",
					},
				},
				ContextCourseRole: model.RoleOwner,
			},
			expected: &model.UserOpResult{
				Email:        "new@test.com",
				SystemErrors: []string{"Users must have a server role to upsert users."},
			},
			expectedUser: nil,
		},

		// Add a user with a higher server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.com",
						Role:  model.GetServerUserRoleString(model.ServerRoleOwner),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:            "new@test.com",
				ValidationErrors: []string{"User has a server role of 'admin', which is not high enough to create a user with server role 'owner'."},
			},
			expectedUser: nil,
		},

		// Add a user with only an insufficient server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.com",
					},
				},
				ContextServerRole: model.ServerRoleUser,
			},
			expected: &model.UserOpResult{
				Email:            "new@test.com",
				ValidationErrors: []string{"User has an insufficient server role of 'user' and no course role to create users."},
			},
			expectedUser: nil,
		},

		// Add a user without proper server or course roles.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "new@test.com",
						Role:       model.GetServerUserRoleString(model.ServerRoleUser),
						Course:     "new-course",
						CourseRole: model.GetCourseUserRoleString(model.RoleStudent),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.RoleStudent,
			},
			expected: &model.UserOpResult{
				Email:            "new@test.com",
				ValidationErrors: []string{"User has a course role of 'student', which is not high enough to create users."},
			},
			expectedUser: nil,
		},

		// Add a user that has a higher course role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "new@test.com",
						Role:       model.GetServerUserRoleString(model.ServerRoleUser),
						Course:     "new-course",
						CourseRole: model.GetCourseUserRoleString(model.RoleOwner),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.RoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:            "new@test.com",
				ValidationErrors: []string{"User has a course role of 'admin', which is not high enough to create a user with course role 'owner'."},
			},
			expectedUser: nil,
		},

		// Validation Errors
		// Most validation errors are already tested by ServerUser (Validation() and Merge()).

		// Update user with course (enroll) without a role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:  "student@test.com",
						Course: "new-course",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:            "student@test.com",
				ValidationErrors: []string{"User 'student@test.com' has an unknown role for course 'new-course'. All users must have a definite role."},
			},
			expectedUser: nil,
		},
	}

	for i, testCase := range testCases {
		// Make sure to test dry runs first, since the expected objects will not be modified.
		success := testUpsertDryRun(test, i, testCase.options, testCase.expected)
		if !success {
			continue
		}

		success = testUpsert(test, i, testCase.options, testCase.expected, testCase.expectedUser)
		if !success {
			continue
		}
	}
}

func testUpsertDryRun(test *testing.T, caseIndex int, options UpsertUsersOptions, expected *model.UserOpResult) bool {
	db.ResetForTesting()

	options.DryRun = true

	beforeUser := db.MustGetServerUser(expected.Email, true)

	result := UpsertUser(options)
	if result == nil {
		test.Errorf("Case (dry run) %d: Got a nil result (which should never happen).", caseIndex)
		return false
	}

	// Fill in the cleartext password if requested (which would be randomly generated).
	if expected.CleartextPassword != "" {
		expected.CleartextPassword = result.CleartextPassword
	}

	if !reflect.DeepEqual(expected, result) {
		test.Errorf("Case (dry run) %d: Result is not as expected. Expected: '%s', Actual: '%s'.", caseIndex,
			util.MustToJSONIndent(expected), util.MustToJSONIndent(result))
		return false
	}

	afterUser := db.MustGetServerUser(expected.Email, true)
	if !reflect.DeepEqual(beforeUser, afterUser) {
		test.Errorf("Case %d: User was changed during a dry run. Before: '%s', After: '%s'.", caseIndex,
			util.MustToJSONIndent(beforeUser), util.MustToJSONIndent(afterUser))
		return false
	}

	return true
}

func testUpsert(test *testing.T, caseIndex int, options UpsertUsersOptions, expected *model.UserOpResult, expectedUser *model.ServerUser) bool {
	db.ResetForTesting()

	options.DryRun = false

	result := UpsertUser(options)
	if result == nil {
		test.Errorf("Case %d: Got a nil result (which should never happen).", caseIndex)
		return false
	}

	// Fill in the cleartext password if requested (which would be randomly generated).
	if expected.CleartextPassword != "" {
		expected.CleartextPassword = result.CleartextPassword
	}

	if !reflect.DeepEqual(expected, result) {
		test.Errorf("Case %d: Result is not as expected. Expected: '%s', Actual: '%s'.", caseIndex,
			util.MustToJSONIndent(expected), util.MustToJSONIndent(result))
		return false
	}

	// If there is an expected user, then check it.
	if expectedUser == nil {
		return true
	}

	actualUser := db.MustGetServerUser(expectedUser.Email, true)

	if actualUser == nil {
		test.Errorf("Case %d: Could not find expected user '%s' in database.", caseIndex, expectedUser.Email)
		return false
	}

	// Do not check salt and tokens exactly, just ensure that the counts match.

	expectedHasSalt := (expectedUser.Salt != nil)
	actualHasSalt := (actualUser.Salt != nil)

	if expectedHasSalt != actualHasSalt {
		test.Errorf("Case %d: Salt not as expected. Expected: '%v', Actual: '%v'.", caseIndex, expectedHasSalt, actualHasSalt)
		return false
	}

	expectedTokenCount := len(expectedUser.Tokens)
	actualTokenCount := len(actualUser.Tokens)

	if expectedTokenCount != actualTokenCount {
		test.Errorf("Case %d: Token count not as expected. Expected: '%d', Actual: '%d'.", caseIndex, expectedTokenCount, actualTokenCount)
		return false
	}

	// After counts have been checked, set the salt and tokens so that the equality check can go through.
	expectedUser.Salt = nil
	actualUser.Salt = nil
	expectedUser.Tokens = nil
	actualUser.Tokens = nil

	// Adjust for the expected user not being validated.

	if expectedUser.Roles == nil {
		expectedUser.Roles = make(map[string]model.CourseUserRole, 0)
	}

	if expectedUser.LMSIDs == nil {
		expectedUser.LMSIDs = make(map[string]string, 0)
	}

	if !reflect.DeepEqual(expectedUser, actualUser) {
		test.Errorf("Case %d: User is not as expected. Expected: '%s', Actual: '%s'.", caseIndex,
			util.MustToJSONIndent(expectedUser), util.MustToJSONIndent(actualUser))
		return false
	}

	return true
}
