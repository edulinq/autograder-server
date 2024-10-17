package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/users"
	"github.com/edulinq/autograder/internal/util"
)

func TestUpsert(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		email     string
		permError bool
		options   *users.UpsertUsersOptions
		expected  []*model.ExternalUserOpResult
	}{
		// Valid permissions.
		// New user without course.
		{
			email:     "server-admin",
			permError: false,
			options: &users.UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "new@test.edulinq.org",
						Name:  "new",
						Role:  model.GetServerUserRoleString(model.ServerRoleUser),
						Pass:  "1234",
					},
				},
				SendEmails: true,
			},
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email:   "new@test.edulinq.org",
						Added:   true,
						Emailed: true,
					},
				},
			},
		},

		// New user with course.
		{
			email:     "server-admin",
			permError: false,
			options: &users.UpsertUsersOptions{
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
				SendEmails: true,
			},
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email:    "new@test.edulinq.org",
						Added:    true,
						Emailed:  true,
						Enrolled: []string{"new-course"},
					},
				},
			},
		},

		// Update user without course.
		{
			email:     "server-owner",
			permError: false,
			options: &users.UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email: "course-student@test.edulinq.org",
						Name:  "new",
					},
				},
				SendEmails: true,
			},
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email:    "course-student@test.edulinq.org",
						Modified: true,
					},
				},
			},
		},

		// Update user with course (enroll).
		{
			email:     "server-owner",
			permError: false,
			options: &users.UpsertUsersOptions{
				RawUsers: []*model.RawServerUserData{
					&model.RawServerUserData{
						Email:       "course-student@test.edulinq.org",
						Course:      "new-course",
						CourseRole:  model.GetCourseUserRoleString(model.CourseRoleStudent),
						CourseLMSID: "new-lms",
					},
				},
				SendEmails: true,
			},
			expected: []*model.ExternalUserOpResult{
				&model.ExternalUserOpResult{
					BaseUserOpResult: model.BaseUserOpResult{
						Email:    "course-student@test.edulinq.org",
						Modified: true,
						Emailed:  true,
						Enrolled: []string{"new-course"},
					},
				},
			},
		},

		// Invalid permissions.
		{"server-user", true, &users.UpsertUsersOptions{}, nil},
		{"server-creator", true, &users.UpsertUsersOptions{}, nil},
		{"course-admin", true, &users.UpsertUsersOptions{}, nil},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		fields := map[string]any{
			"raw-users":    testCase.options.RawUsers,
			"skip-inserts": testCase.options.SkipInserts,
			"skip-updates": testCase.options.SkipUpdates,
			"dry-run":      testCase.options.DryRun,
			"send-emails":  testCase.options.SendEmails,
		}

		response := core.SendTestAPIRequestFull(test, `users/upsert`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-041"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, expectedLocator, response.Locator)
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

		var responseContent UpsertResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(testCase.expected, responseContent.Results) {
			test.Errorf("Case %d: Unexpected user op response. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent.Results))
			continue
		}
	}
}
