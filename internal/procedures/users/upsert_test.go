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
var PERMISSION_ERROR_EXTERNAL_MESSAGE = "You have insufficient permissions for the requested operation."

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
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "new@test.edulinq.org",
						Name:  "new",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
						Pass:  "1234",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:   "new@test.edulinq.org",
					Added:   true,
					Emailed: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "new@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
			},
		},

		// New user with course.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
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
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "new@test.edulinq.org",
					Added:    true,
					Emailed:  true,
					Enrolled: []string{"new-course"},
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "new@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
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
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "course-student@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "course-student@test.edulinq.org",
					Modified: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
				},
			},
		},

		// Update user with course (enroll).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email:       "course-student@test.edulinq.org",
						Course:      "new-course",
						CourseRole:  model.GetCourseUserRoleString(model.CourseRoleStudent),
						CourseLMSID: "new-lms",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "course-student@test.edulinq.org",
					Modified: true,
					Emailed:  true,
					Enrolled: []string{"new-course"},
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
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
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email:      "course-student@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleGrader),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "course-student@test.edulinq.org",
					Modified: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleGrader,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
				},
			},
		},

		// Update course information using course (not server) permissions.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email:      "course-student@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleGrader),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "course-student@test.edulinq.org",
					Modified: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleGrader,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
				},
			},
		},

		// Update self with non-admin permissions.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "course-student@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextEmail:      "course-student@test.edulinq.org",
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleStudent,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "course-student@test.edulinq.org",
					Modified: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
				},
			},
		},

		// Update self with non-admin permissions and no course.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "course-student@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextEmail:      "course-student@test.edulinq.org",
				ContextServerRole: model.ServerRoleUser,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "course-student@test.edulinq.org",
					Modified: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("new"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
				},
			},
		},

		// Self demote (server).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "server-admin@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextEmail:      "server-admin@test.edulinq.org",
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "server-admin@test.edulinq.org",
					Modified: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "server-admin@test.edulinq.org",
				Name:     util.StringPointer("server-admin"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
			},
		},

		// Self demote (course).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
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
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "course-admin@test.edulinq.org",
					Modified: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-admin@test.edulinq.org",
				Name:     util.StringPointer("course-admin"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleAdmin,
						LMSID: util.StringPointer("lms-course-admin@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-admin@test.edulinq.org"),
					},
				},
			},
		},

		// Set a new password.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "course-student@test.edulinq.org",
						Pass:  "1234",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "course-student@test.edulinq.org",
					Modified: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
				},
			},
		},

		// Set a duplicate password.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "course-student@test.edulinq.org",
						Pass:  util.Sha256HexFromString("course-student"),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "course-student@test.edulinq.org",
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-student@test.edulinq.org",
				Name:     util.StringPointer("course-student"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleStudent,
						LMSID: util.StringPointer("lms-course-student@test.edulinq.org"),
					},
				},
			},
		},

		// New user without password.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "new@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:   "new@test.edulinq.org",
					Added:   true,
					Emailed: true,
				},
				CleartextPassword: PLACEHOLDER_PASSWORD_CLEARTEXT,
			},
			expectedUser: &model.ServerUser{
				Email:    "new@test.edulinq.org",
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
			},
		},

		// Add a user from a course (but not server) admin.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "new@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:   "new@test.edulinq.org",
					Added:   true,
					Emailed: true,
				},
				CleartextPassword: PLACEHOLDER_PASSWORD_CLEARTEXT,
			},
			expectedUser: &model.ServerUser{
				Email:    "new@test.edulinq.org",
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
			},
		},

		// Update course information on a user with higher server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email:      "course-owner@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleAdmin),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email:    "course-owner@test.edulinq.org",
					Modified: true,
				},
			},
			expectedUser: &model.ServerUser{
				Email:    "course-owner@test.edulinq.org",
				Name:     util.StringPointer("course-owner"),
				Salt:     PLACEHOLDER_SALT,
				Password: PLACEHOLDER_PASSWORD_TOKEN,
				Role:     model.ServerRoleUser,
				CourseInfo: map[string]*model.UserCourseInfo{
					"course-languages": &model.UserCourseInfo{
						Role:  model.CourseRoleOwner,
						LMSID: util.StringPointer("lms-course-owner@test.edulinq.org"),
					},
					"course101": &model.UserCourseInfo{
						Role:  model.CourseRoleAdmin,
						LMSID: util.StringPointer("lms-course-owner@test.edulinq.org"),
					},
				},
			},
		},

		// Permission Errors

		// Add a user wihout a any roles.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "new@test.edulinq.org",
					},
				},
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "",
				},
				SystemError: &model.LocatableError{
					Locator:         "-1001",
					HideLocator:     false,
					InternalMessage: "No authority/roles were provided when adding a user.",
					ExternalMessage: "The server failed to process your request. Please contact an administrator with this ID '-1001'.",
				},
			},
			expectedUser: nil,
		},

		// Add a user with only a course role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "new@test.edulinq.org",
					},
				},
				ContextCourseRole: model.CourseRoleOwner,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "new@test.edulinq.org",
				},
				SystemError: &model.LocatableError{
					Locator:         "-1010",
					HideLocator:     false,
					InternalMessage: "Users must have a server role to upsert users.",
					ExternalMessage: "The server failed to process your request. Please contact an administrator with this ID '-1010'.",
				},
			},
			expectedUser: nil,
		},

		// Add a user with a higher server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "new@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleOwner),
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "new@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1011",
					HideLocator:     true,
					InternalMessage: "User has a server role of 'admin', which is not high enough to upsert a user with server role of 'owner'.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Add a user with only an insufficient server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "new@test.edulinq.org",
					},
				},
				ContextServerRole: model.ServerRoleUser,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "new@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1012",
					HideLocator:     true,
					InternalMessage: "User has an insufficient server role of 'user' and no course role to insert users.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Add a user without proper server or course roles.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
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
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "new@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1014",
					HideLocator:     true,
					InternalMessage: "User has a course role of 'student', which is not high enough to insert users.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Add a user that has a higher course role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
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
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "new@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1013",
					HideLocator:     true,
					InternalMessage: "User has a course role of 'admin', which is not high enough to insert a user with course role of 'owner'.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Update a user to have a higher course role than the context user.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email:      "course-grader@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleOwner),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "course-grader@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1022",
					HideLocator:     true,
					InternalMessage: "User has a course role of 'admin', which is not high enough to update a user to a course role of 'owner'.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Update server information on a user that has a higher server role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "server-owner@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "server-owner@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1016",
					HideLocator:     true,
					InternalMessage: "User has a server role of 'admin', which is not high enough to update a user with server role of 'owner'.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Update server role to a higher server role than the context user.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "server-user@test.edulinq.org",
						Role:  "owner",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "server-user@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1011",
					HideLocator:     true,
					InternalMessage: "User has a server role of 'admin', which is not high enough to upsert a user with server role of 'owner'.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Update course information on a user that has a higher course role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email:      "course-owner@test.edulinq.org",
						Course:     "course101",
						CourseRole: model.GetCourseUserRoleString(model.CourseRoleGrader),
					},
				},
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "course-owner@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1019",
					HideLocator:     true,
					InternalMessage: "User has a course role of 'admin', which is not high enough to update a user with course role of 'owner'.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Update non-self with non-admin permissions (server).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "course-student@test.edulinq.org",
						Name:  "new",
					},
				},
				ContextEmail:      "ZZZ@test.edulinq.org",
				ContextServerRole: model.ServerRoleUser,
				ContextCourseRole: model.CourseRoleStudent,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "course-student@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1017",
					HideLocator:     true,
					InternalMessage: "User has a server role of 'user', which is not high enough to update server-level information for another user.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Update non-self with non-admin permissions (course).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
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
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "course-student@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1020",
					HideLocator:     true,
					InternalMessage: "User has a course role of 'student', which is not high enough to update course-level information for another user.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Cannot self promote (server).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "course-admin@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleOwner),
					},
				},
				ContextEmail:      "course-admin@test.edulinq.org",
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "course-admin@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1011",
					HideLocator:     true,
					InternalMessage: "User has a server role of 'admin', which is not high enough to upsert a user with server role of 'owner'.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Cannot self promote (course).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
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
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "course-grader@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1018",
					HideLocator:     true,
					InternalMessage: "User has a course role of 'grader', which is not high enough to modify course roles.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Non-admins cannot self demote (server).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "server-creator@test.edulinq.org",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
					},
				},
				ContextServerRole: model.ServerRoleCourseCreator,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "server-creator@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1015",
					HideLocator:     true,
					InternalMessage: "User has a server role of 'creator', which is not high enough to modify server roles.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Non-admins cannot self demote (course).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
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
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "course-grader@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1018",
					HideLocator:     true,
					InternalMessage: "User has a course role of 'grader', which is not high enough to modify course roles.",
					ExternalMessage: PERMISSION_ERROR_EXTERNAL_MESSAGE,
				},
			},
			expectedUser: nil,
		},

		// Validation Errors
		// Most validation errors are already tested by ServerUser (Validation() and Merge()).

		// Update user with course (enroll) without a role.
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email:  "course-student@test.edulinq.org",
						Course: "new-course",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "course-student@test.edulinq.org",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1003",
					HideLocator:     true,
					InternalMessage: "User 'course-student@test.edulinq.org' has an invalid course info for course 'new-course': 'Unknown course role.'.",
					ExternalMessage: "User 'course-student@test.edulinq.org' has an invalid course info for course 'new-course': 'Unknown course role.'.",
				},
			},
			expectedUser: nil,
		},

		// Insert a user that fails to validate (bad name).
		{
			options: UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "new-user",
					},
				},
				ContextServerRole: model.ServerRoleAdmin,
			},
			expected: &model.UserOpResult{
				BaseUserOpResult: model.BaseUserOpResult{
					Email: "new-user",
				},
				ValidationError: &model.LocatableError{
					Locator:         "-1003",
					HideLocator:     true,
					InternalMessage: "User email 'new-user' has an invalid format.",
					ExternalMessage: "User email 'new-user' has an invalid format.",
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

	return user.Clone()
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

	if !reflect.DeepEqual(expected, result) {
		test.Errorf("Case (dry run, email: %v) %d: Result is not as expected. Expected: '%s', Actual: '%s'.",
			sendEmails, caseIndex, util.MustToJSONIndent(expected), util.MustToJSONIndent(result))
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

	if !reflect.DeepEqual(expected, result) {
		test.Errorf("Case (wet run, email: %v) %d: Result is not as expected. Expected: '%s', Actual: '%s'.",
			sendEmails, caseIndex, util.MustToJSONIndent(expected), util.MustToJSONIndent(result))
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

	// Do not check salt, and password exactly.
	// Ignore tokens.

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
