package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

var PLACEHOLDER_PASSWORD_CLEARTEXT string = "<placeholder_pass>"
var PLACEHOLDER_SALT *string = util.StringPointer("abcd")
var PLACEHOLDER_PASSWORD_TOKEN *model.Token = model.MustNewToken(PLACEHOLDER_PASSWORD_CLEARTEXT, *PLACEHOLDER_SALT, model.TokenSourcePassword, "password")
var VALIDATION_ERROR_EXTERNAL_MESSAGE = "You have insufficient permissions for the requested operation."

func TestUpsertUser(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	email.ClearTestMessages()
	defer email.ClearTestMessages()

	testCases := []struct {
		// Do not set DryRun or SendEmails, these will be set automatically.
		options      UpsertUsersOptions
		expected     *model.UserOpResult
		expectedUser *model.ServerUser
	}{
		// New user without course.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.edulinq.org",
						Name:  "new",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
						Pass:  "1234",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:   "new@test.edulinq.org",
				Added:   true,
				Emailed: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "new@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
			},
		},

		// New user with course.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:       "new@test.edulinq.org",
						Name:        "new",
						Role:        model.GetServerUserRoleString(model.ServerRoleUser),
						Pass:        "1234",
						Course:      "new-course",
						CourseRole:  model.GetCourseUserRoleString(model.CourseRoleStudent),
						CourseLMSID: "new-lms",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "new@test.edulinq.org",
				Added:    true,
				Emailed:  true,
				Enrolled: []string{"new-course"},
			},
			expectedUser: &model.ServerUser{
				Email:    "new@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"new-course": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("new-lms"),
					},
				},
			},
		},

		// Update user without course.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "course-student@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "course-student@test.edulinq.org",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
				},
			},
		},

		// Update user with course (enroll).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:       "course-student@test.edulinq.org",
						Course:      "new-course",
						CourseRole:  model.GetCourseUserRoleString(model.CourseRoleStudent),
						CourseLMSID: "new-lms",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "course-student@test.edulinq.org",
				Modified: true,
				Emailed:  true,
				Enrolled: []string{"new-course"},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
					"new-course": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("new-lms"),
					},
				},
			},
		},

		// Update course information.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "course-student@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleGrader),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "course-student@test.edulinq.org",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleGrader,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
				},
			},
		},

		// Update course information using course (not server) permissions.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "course-student@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleGrader),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "course-student@test.edulinq.org",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleGrader,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
				},
			},
		},

		// Update self with non-admin permissions.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "course-student@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextEmail:      "course-student@test.edulinq.org",
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleStudent,
			},
			expected: &model.UserOpResult{
				Email:    "course-student@test.edulinq.org",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
				},
			},
		},

		// Update self with non-admin permissions and no course.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "course-student@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextEmail:      "course-student@test.edulinq.org",
				ContextServerRole: model.ServerRoleUser,
			},
			expected: &model.UserOpResult{
				Email:    "course-student@test.edulinq.org",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
				},
			},
		},

		// Self demote (server).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "server-admin@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextEmail:      "server-admin@test.edulinq.org",
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "server-admin@test.edulinq.org",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "server-admin@test.edulinq.org",
				Name:     util.StringPointer("server-admin"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
			},
		},

		// Self demote (course).
		// TODO: This test does not fully test the functionality. We need a server non-admin and a course admin.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "course-admin@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleStudent),
					},
				},
				ContextEmail:      "course-admin@test.edulinq.org",
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "course-admin@test.edulinq.org",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "course-admin@test.edulinq.org",
				Name:     util.StringPointer("course-admin"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleAdmin,
						LMSID: util.StringPointer("lms-course-admin@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleAdmin,
						LMSID: util.StringPointer("lms-course-admin@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleAdmin,
						LMSID: util.StringPointer("lms-course-admin@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleAdmin,
					},
				},
			},
		},

		// Set a new password.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "course-student@test.edulinq.org",
						Pass:  "1234",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "course-student@test.edulinq.org",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
				},
			},
		},

		// Set a duplicate password.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "course-student@test.edulinq.org",
						Pass:  util.Sha256HexFromString("course-student"),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email: "course-student@test.edulinq.org",
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleStudent,
					},
				},
			},
		},

		// New user without password.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:             "new@test.edulinq.org",
				Added:             true,
				Emailed:           true,
				CleartextPassword: PLACEHOLDER_PASSWORD_CLEARTEXT,
			},
			expectedUser: &model.ServerUser{
				Email:    "new@test.edulinq.org",
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
			},
		},

		// Add a user from a course (but not server) admin.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:             "new@test.edulinq.org",
				Added:             true,
				Emailed:           true,
				CleartextPassword: PLACEHOLDER_PASSWORD_CLEARTEXT,
			},
			expectedUser: &model.ServerUser{
				Email:    "new@test.edulinq.org",
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
			},
		},

		// Update course information on a user with higher server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "course-owner@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleAdmin),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email:    "course-owner@test.edulinq.org",
				Modified: true,
			},
			expectedUser: &model.ServerUser{
				Email:    "course-owner@test.edulinq.org",
				Name:     util.StringPointer("course-owner"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				Tokens:   []*model.Token{},
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleOwner,
						LMSID: util.StringPointer("lms-course-owner@test.edulinq.org"),
					},
					"course-with-lms": &model.UserCourseInfo{
						Role:  model.CourseRoleOwner,
						LMSID: util.StringPointer("lms-course-owner@test.edulinq.org"),
					},
					"course-without-source": &model.UserCourseInfo{
						Role:  model.CourseRoleOwner,
						LMSID: util.StringPointer("lms-course-owner@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role: model.CourseRoleAdmin,
					},
					"course101-with-zero-limit": &model.UserCourseInfo{
						Role: model.CourseRoleOwner,
					},
				},
			},
		},

		// Permission Errors

		// Add a user wihout a any roles.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.edulinq.org",
					},
				},
			},
			expected: &model.UserOpResult{
				Email: "",
				SystemErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1001",
						InternalMessage: "No authority/roles were provided when adding a user.",
						ExternalMessage: "The server failed to process your request. Please contact an administrator with this ID '-1001'.",
					},
				},
			},
			expectedUser: nil,
		},

		// Add a user with only a course role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.edulinq.org",
					},
				},
				ContextCourseRole: model.CourseRoleOwner,
			},
			expected: &model.UserOpResult{
				Email: "new@test.edulinq.org",
				SystemErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1010",
						InternalMessage: "Users must have a server role to upsert users.",
						ExternalMessage: "The server failed to process your request. Please contact an administrator with this ID '-1010'.",
					},
				},
			},
			expectedUser: nil,
		},

		// Add a user with a higher server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleOwner),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email: "new@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1011",
						InternalMessage: "User has a server role of 'admin', which is not high enough to upsert a user with server role of 'owner'.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Add a user with only an insufficient server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "new@test.edulinq.org",
					},
				},
				ContextServerRole: model.ServerRoleUser,
			},
			expected: &model.UserOpResult{
				Email: "new@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1012",
						InternalMessage: "User has an insufficient server role of 'user' and no course role to insert users.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Add a user without proper server or course roles.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "new@test.edulinq.org",
						Role:       model.GetServerUserRoleString(model.ServerRoleUser),
						Course:     "new-course",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleStudent),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleStudent,
			},
			expected: &model.UserOpResult{
				Email: "new@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1014",
						InternalMessage: "User has a course role of 'student', which is not high enough to insert users.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Add a user that has a higher course role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "new@test.edulinq.org",
						Role:       model.GetServerUserRoleString(model.ServerRoleUser),
						Course:     "new-course",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleOwner),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email: "new@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1013",
						InternalMessage: "User has a course role of 'admin', which is not high enough to insert a user with course role of 'owner'.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Update server information on a user that has a higher server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "server-owner@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email: "server-owner@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1016",
						InternalMessage: "User has a server role of 'admin', which is not high enough to update a user with server role of 'owner'.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Update course information on a user that has a higher course role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "course-owner@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleGrader),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email: "course-owner@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1019",
						InternalMessage: "User has a course role of 'admin', which is not high enough to update a user with course role of 'owner'.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Update non-self with non-admin permissions (server).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "course-student@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextEmail:      "ZZZ@test.edulinq.org",
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleStudent,
			},
			expected: &model.UserOpResult{
				Email: "course-student@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1017",
						InternalMessage: "User has a server role of 'user', which is not high enough to update server-level information for another user.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Update non-self with non-admin permissions (course).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:       "course-student@test.edulinq.org",
						Course:      "course101",
						CourseRole:  model.GetCourseUserRoleString(model.CourseRoleStudent),
						CourseLMSID: "new",
					},
				},
				ContextEmail:      "ZZZ@test.edulinq.org",
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleStudent,
			},
			expected: &model.UserOpResult{
				Email: "course-student@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1020",
						InternalMessage: "User has a course role of 'student', which is not high enough to update course-level information for another user.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Cannot self promote (server).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "course-admin@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleOwner),
					},
				},
				ContextEmail:      "course-admin@test.edulinq.org",
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email: "course-admin@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1011",
						InternalMessage: "User has a server role of 'admin', which is not high enough to upsert a user with server role of 'owner'.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Cannot self promote (course).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "course-grader@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleAdmin),
					},
				},
				ContextEmail:      "course-grader@test.edulinq.org",
				ContextServerRole: model.ServerRoleCourseCreator,
				ContextCourseRole: model.CourseRoleGrader,
			},
			expected: &model.UserOpResult{
				Email: "course-grader@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1018",
						InternalMessage: "User has a course role of 'grader', which is not high enough to modify course roles.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Non-admins cannot self demote (server).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email: "server-creator@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextServerRole: model.ServerRoleCourseCreator,
			},
			expected: &model.UserOpResult{
				Email: "server-creator@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1015",
						InternalMessage: "User has a server role of 'creator', which is not high enough to modify server roles.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},

		// Non-admins cannot self demote (course).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawUserData{
					&model.RawUserData{
						Email:      "course-grader@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleStudent),
					},
				},
				ContextEmail:      "course-grader@test.edulinq.org",
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleGrader,
			},
			expected: &model.UserOpResult{
				Email: "course-grader@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1018",
						InternalMessage: "User has a course role of 'grader', which is not high enough to modify course roles.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
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
						Email:  "course-student@test.edulinq.org",
						Course: "new-course",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				Email: "course-student@test.edulinq.org",
				ValidationErrors: []*model.ModelError{
					&model.ModelError{
						Locator:         "-1003",
						InternalMessage: "User 'course-student@test.edulinq.org' has an invalid course info 'new-course': 'Unknown course role.'.",
						ExternalMessage: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
			expectedUser: nil,
		},
	}

	for i, testCase := range testCases {
		// Dry run without emails.
		success := testUpsertDryRun(test, i, false, testCase.options, testCase.expected.MustClone())
		if !success {
			continue
		}

		// Dry run with emails.
		success = testUpsertDryRun(test, i, true, testCase.options, testCase.expected.MustClone())
		if !success {
			continue
		}

		// Wet run without emails.
		success = testUpsert(test, i, false, testCase.options, testCase.expected.MustClone(), cloneTestServerUser(testCase.expectedUser))
		if !success {
			continue
		}

		// Wet run with emails.
		success = testUpsert(test, i, true, testCase.options, testCase.expected.MustClone(), cloneTestServerUser(testCase.expectedUser))
		if !success {
			continue
		}
	}
}

func cloneTestServerUser(user *model.ServerUser) *model.ServerUser {
	if user == nil {
		return nil
	}

	// Specially copy the tokens (since we will be passing nils instead of real tokens).
	newTokens := append([]*model.Token(nil), user.Tokens...)
	oldTokens := user.Tokens

	user.Tokens = nil
	clone := user.Clone()

	user.Tokens = oldTokens
	clone.Tokens = newTokens

	return clone
}

func testUpsertDryRun(test *testing.T, caseIndex int, sendEmails bool, options UpsertUsersOptions, expected *model.UserOpResult) bool {
	db.ResetForTesting()
	email.ClearTestMessages()

	options.DryRun = true
	options.SendEmails = sendEmails

	expected.Emailed = expected.Emailed && sendEmails

	beforeUser := db.MustGetServerUser(expected.Email)

	result := UpsertUser(options)
	if result == nil {
		test.Errorf("Case (dry run, email: %v) %d: Got a nil result (which should never happen).", sendEmails, caseIndex)
		return false
	}

	// Fill in the cleartext password if requested (which would be randomly generated).
	if expected.CleartextPassword != "" {
		expected.CleartextPassword = result.CleartextPassword
	}

	if !model.ModelErrorSlicesEquals(expected.ValidationErrors, result.ValidationErrors) {
		expectedValidationErrors := model.DereferenceModelErrors(expected.ValidationErrors)
		actualValidationErrors := model.DereferenceModelErrors(result.ValidationErrors)

		test.Errorf("Case (dry run, email: %v) %d: Validation Errors are not as expected. Expected: '%+v', actual: '%+v'.",
			sendEmails, caseIndex, expectedValidationErrors, actualValidationErrors)
		return false
	}

	if !model.ModelErrorSlicesEquals(expected.SystemErrors, result.SystemErrors) {
		expectedSystemErrors := model.DereferenceModelErrors(expected.SystemErrors)
		actualSystemErrors := model.DereferenceModelErrors(result.SystemErrors)

		test.Errorf("Case (dry run, email: %v) %d: System Errors are not as expected. Expected: '%+v', actual: '%+v'.",
			sendEmails, caseIndex, expectedSystemErrors, actualSystemErrors)
		return false
	}

	// Clear expected errors so we can use reflection.
	expected.ValidationErrors = []*model.ModelError{}
	result.ValidationErrors = []*model.ModelError{}
	expected.SystemErrors = []*model.ModelError{}
	result.SystemErrors = []*model.ModelError{}

	if !reflect.DeepEqual(expected, result) {
		test.Errorf("Case (dry run, email: %v) %d: Result is not as expected. Expected: '%+v', Actual: '%+v'.",
			sendEmails, caseIndex, expected, result)
		return false
	}

	afterUser := db.MustGetServerUser(expected.Email)
	if !reflect.DeepEqual(beforeUser, afterUser) {
		test.Errorf("Case (dry run, email: %v) %d: User was changed during a dry run. Before: '%s', After: '%s'.", sendEmails, caseIndex,
			util.MustToJSONIndent(beforeUser), util.MustToJSONIndent(afterUser))
		return false
	}

	// Ensure that no emails are ever sent.
	emailCount := len(email.GetTestMessages())
	if emailCount > 0 {
		test.Errorf("Case (dry run, email: %v) %d: %d emails were sent, when none should be sent on a dry run.", sendEmails, caseIndex, emailCount)
		return false
	}

	return true
}

func testUpsert(test *testing.T, caseIndex int, sendEmails bool, options UpsertUsersOptions, expected *model.UserOpResult, expectedUser *model.ServerUser) bool {
	db.ResetForTesting()
	email.ClearTestMessages()

	options.DryRun = false
	options.SendEmails = sendEmails

	expected.Emailed = expected.Emailed && sendEmails

	result := UpsertUser(options)
	if result == nil {
		test.Errorf("Case (wet run, email: %v) %d: Got a nil result (which should never happen).", sendEmails, caseIndex)
		return false
	}

	// Fill in the cleartext password if requested (which would be randomly generated).
	if expected.CleartextPassword != "" {
		expected.CleartextPassword = result.CleartextPassword
	}

	if !model.ModelErrorSlicesEquals(expected.ValidationErrors, result.ValidationErrors) {
		expectedValidationErrors := model.DereferenceModelErrors(expected.ValidationErrors)
		actualValidationErrors := model.DereferenceModelErrors(result.ValidationErrors)

		test.Errorf("Case (wet run, email: %v) %d: Validation Errors are not as expected. Expected: '%+v', actual: '%+v'.",
			sendEmails, caseIndex, expectedValidationErrors, actualValidationErrors)
		return false
	}

	if !model.ModelErrorSlicesEquals(expected.SystemErrors, result.SystemErrors) {
		expectedSystemErrors := model.DereferenceModelErrors(expected.SystemErrors)
		actualSystemErrors := model.DereferenceModelErrors(result.SystemErrors)

		test.Errorf("Case (wet run, email: %v) %d: System Errors are not as expected. Expected: '%+v', actual: '%+v'.",
			sendEmails, caseIndex, expectedSystemErrors, actualSystemErrors)
		return false
	}

	// Clear expected errors so we can use reflection.
	expected.ValidationErrors = []*model.ModelError{}
	result.ValidationErrors = []*model.ModelError{}
	expected.SystemErrors = []*model.ModelError{}
	result.SystemErrors = []*model.ModelError{}

	if !reflect.DeepEqual(expected, result) {
		test.Errorf("Case (wet run, email: %v) %d: Result is not as expected. Expected: '%+v', Actual: '%+v'.",
			sendEmails, caseIndex, expected, result)
		return false
	}

	// If there is an expected user, then check it.
	if expectedUser == nil {
		return true
	}

	actualUser := db.MustGetServerUser(expectedUser.Email)

	if actualUser == nil {
		test.Errorf("Case (wet run, email: %v) %d: Could not find expected user '%s' in database.", sendEmails, caseIndex, expectedUser.Email)
		return false
	}

	// Do not check salt, password, and tokens exactly, just ensure that the counts match.

	expectedHasSalt := (expectedUser.Salt != nil)
	actualHasSalt := (actualUser.Salt != nil)
	if expectedHasSalt != actualHasSalt {
		test.Errorf("Case (wet run, email: %v) %d: Salt not as expected. Expected: '%v', Actual: '%v'.", sendEmails, caseIndex, expectedHasSalt, actualHasSalt)
		return false
	}

	expectedHasPassword := (expectedUser.Password != nil)
	actualHasPassword := (actualUser.Password != nil)
	if expectedHasPassword != actualHasPassword {
		test.Errorf("Case (wet run, email: %v) %d: Password not as expected. Expected: '%v', Actual: '%v'.", sendEmails, caseIndex, expectedHasPassword, actualHasPassword)
		return false
	}

	expectedTokenCount := len(expectedUser.Tokens)
	actualTokenCount := len(actualUser.Tokens)

	if expectedTokenCount != actualTokenCount {
		test.Errorf("Case (wet run, email: %v) %d: Token count not as expected. Expected: '%d', Actual: '%d'.", sendEmails, caseIndex, expectedTokenCount, actualTokenCount)
		return false
	}

	// After counts have been checked, set the salt, password, and tokens so that the equality check can go through.
	expectedUser.Salt = nil
	actualUser.Salt = nil
	expectedUser.Password = nil
	actualUser.Password = nil
	expectedUser.Tokens = nil
	actualUser.Tokens = nil

	// Adjust for the expected user not being validated.

	if expectedUser.CourseInfo == nil {
		expectedUser.CourseInfo = make(map[string]*model.UserCourseInfo, 0)
	}

	if !reflect.DeepEqual(expectedUser, actualUser) {
		test.Errorf("Case (wet run, email: %v) %d: User is not as expected. Expected: '%s', Actual: '%s'.", sendEmails, caseIndex,
			util.MustToJSONIndent(expectedUser), util.MustToJSONIndent(actualUser))
		return false
	}

	// Check that an email was sent.
	emailCount := len(email.GetTestMessages())
	if expected.Emailed && (emailCount != 1) {
		test.Errorf("Case (wet run, email: %v) %d: Expected exactly 1 email to be sent, found %d emails sent.", sendEmails, caseIndex, emailCount)
		return false
	}

	return true
}
