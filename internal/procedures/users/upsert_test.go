package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// TEST - More expected users

const (
	PASS_PLACEHOLDER = "<placeholder_pass>"
)

// TEST - Add tokens

func TestUpsertUser(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
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
				Salt:   util.StringPointer("abcd"),
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
				Salt:   util.StringPointer("abcd"),
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
				Salt:   util.StringPointer("abcd"),
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
				Salt:   util.StringPointer("abcd"),
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
				CleartextPassword: PASS_PLACEHOLDER,
			},
			expectedUser: &model.ServerUser{
				Email:  "new@test.com",
				Salt:   util.StringPointer("abcd"),
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
				CleartextPassword: PASS_PLACEHOLDER,
			},
			expectedUser: &model.ServerUser{
				Email:  "new@test.com",
				Salt:   util.StringPointer("abcd"),
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
		db.ResetForTesting()

		result := UpsertUser(testCase.options)

		// Fill in the cleartext password if requested (which would be randomly generated).
		if testCase.expected.CleartextPassword == PASS_PLACEHOLDER {
			testCase.expected.CleartextPassword = result.CleartextPassword
		}

		if !reflect.DeepEqual(testCase.expected, result) {
			test.Errorf("Case %d: Result is not as expected. Expected: '%s', Actual: '%s'.", i,
				util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(result))
			continue
		}

		// If there is an expected user, then check it.
		if testCase.expectedUser == nil {
			continue
		}

		expectedUser := testCase.expectedUser
		actualUser := db.MustGetServerUser(expectedUser.Email, true)

		if actualUser == nil {
			test.Errorf("Case %d: Could not find expected user '%s' in database.", i, expectedUser.Email)
			continue
		}

		// Do not check salt and tokens exactly, just ensure that the counts match.

		expectedHasSalt := (expectedUser.Salt != nil)
		actualHasSalt := (actualUser.Salt != nil)

		if expectedHasSalt != actualHasSalt {
			test.Errorf("Case %d: Salt not as expected. Expected: '%v', Actual: '%v'.", i, expectedHasSalt, actualHasSalt)
			continue
		}

		expectedTokenCount := len(expectedUser.Tokens)
		actualTokenCount := len(actualUser.Tokens)

		if expectedTokenCount != actualTokenCount {
			test.Errorf("Case %d: Token count not as expected. Expected: '%d', Actual: '%d'.", i, expectedTokenCount, actualTokenCount)
			continue
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
			test.Errorf("Case %d: User is not as expected. Expected: '%s', Actual: '%s'.", i,
				util.MustToJSONIndent(expectedUser), util.MustToJSONIndent(actualUser))
			continue
		}

	}
}
