package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

var VALIDATION_ERROR_EXTERNAL_MESSAGE = "You have insufficient permissions for the requested operation."

func TestEnroll(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		email          string
		permError      bool
		locator        string
		rawCourseUsers []*model.RawCourseUserData
		sendEmails     bool
		skipInserts    bool
		skipUpdates    bool
		dryRun         bool
		expected       []*model.ExternalUserOpResult
	}{
		// Valid permissions.
		// New user.
		{
			email:     "course-admin",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email:       "new@test.edulinq.org",
					Name:        "new",
					CourseRole:  model.GetCourseUserRoleString(model.CourseRoleStudent),
					CourseLMSID: "new-lms",
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email:    "new@test.edulinq.org",
						Added:    true,
						Emailed:  true,
						Enrolled: []string{"course101"},
					},
				},
			},
		},

		// Update user course role.
		{
			email:     "course-owner",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email:      "course-student@test.edulinq.org",
					CourseRole: model.GetCourseUserRoleString(model.CourseRoleGrader),
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email:    "course-student@test.edulinq.org",
						Modified: true,
					},
				},
			},
		},

		// Update user lms with course.
		{
			email:     "course-owner",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email:       "course-student@test.edulinq.org",
					CourseRole:  model.GetCourseUserRoleString(model.CourseRoleStudent),
					CourseLMSID: "new-lms",
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email:    "course-student@test.edulinq.org",
						Modified: true,
					},
				},
			},
		},

		// Valid permissions, role escalation.
		// New user with course.
		{
			email:     "server-admin",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email:       "new@test.edulinq.org",
					Name:        "new",
					CourseRole:  model.GetCourseUserRoleString(model.CourseRoleStudent),
					CourseLMSID: "new-lms",
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email:    "new@test.edulinq.org",
						Added:    true,
						Emailed:  true,
						Enrolled: []string{"course101"},
					},
				},
			},
		},

		// Update user with course.
		{
			email:     "server-admin",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email:       "course-student@test.edulinq.org",
					CourseRole:  model.GetCourseUserRoleString(model.CourseRoleStudent),
					CourseLMSID: "new-lms",
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email:    "course-student@test.edulinq.org",
						Modified: true,
					},
				},
			},
		},

		// Invalid permissions, procedure validation error.
		// New user without course.
		{
			email:     "course-admin",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email: "new@test.edulinq.org",
					Name:  "new",
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email: "new@test.edulinq.org",
					},
					ValidationError: &model.ExternalLocatableError{
						Locator: "",
						Message: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
		},

		// Promote existing user to a higher course role.
		{
			email:     "course-admin",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email:      "course-student@test.edulinq.org",
					CourseRole: model.GetCourseUserRoleString(model.CourseRoleOwner),
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email: "course-student@test.edulinq.org",
					},
					ValidationError: &model.ExternalLocatableError{
						Locator: "",
						Message: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
		},

		// Demote existing user with a higher course role.
		{
			email:     "course-admin",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email:      "course-owner@test.edulinq.org",
					CourseRole: model.GetCourseUserRoleString(model.CourseRoleStudent),
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email: "course-owner@test.edulinq.org",
					},
					ValidationError: &model.ExternalLocatableError{
						Locator: "",
						Message: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
		},

		// Update user server info without course.
		{
			email:     "course-owner",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email: "course-student@test.edulinq.org",
					Name:  "new",
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email: "course-student@test.edulinq.org",
					},
					ValidationError: &model.ExternalLocatableError{
						Locator: "",
						Message: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
		},

		// Update user server info, name, with course (enroll).
		{
			email:     "course-owner",
			permError: false,
			locator:   "",
			rawCourseUsers: []*model.RawCourseUserData{
				&model.RawCourseUserData{
					Email:      "course-student@test.edulinq.org",
					Name:       "new",
					CourseRole: model.GetCourseUserRoleString(model.CourseRoleStudent),
				},
			},
			sendEmails:  true,
			skipInserts: false,
			skipUpdates: false,
			dryRun:      false,
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email: "course-student@test.edulinq.org",
					},
					ValidationError: &model.ExternalLocatableError{
						Locator: "",
						Message: VALIDATION_ERROR_EXTERNAL_MESSAGE,
					},
				},
			},
		},

		// Invalid permissions.
		{"course-grader", true, "-020", []*model.RawCourseUserData{}, false, false, false, false, nil},
		{"course-student", true, "-020", []*model.RawCourseUserData{}, false, false, false, false, nil},
		// Invalid permissions, role escalation.
		{"server-creator", true, "-040", []*model.RawCourseUserData{}, false, false, false, false, nil},
		{"server-user", true, "-040", []*model.RawCourseUserData{}, false, false, false, false, nil},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		fields := map[string]any{
			"raw-course-users": testCase.rawCourseUsers,
			"skip-inserts":     testCase.skipInserts,
			"skip-updates":     testCase.skipUpdates,
			"dry-run":          testCase.dryRun,
			"send-emails":      testCase.sendEmails,
		}

		response := core.SendTestAPIRequestFull(test, core.makeFullAPIPath(`courses/users/enroll`), fields, nil, testCase.email)
		if !response.Success {
			if testCase.permError {
				if response.Locator != testCase.locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.locator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.permError {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		var responseContent EnrollResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(testCase.expected, responseContent.Results) {
			test.Errorf("Case %d: Unexpected user op response. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent.Results))
			continue
		}
	}
}
