package user

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// Many of the semantics of add users are tested at the course level,
// focus on the API here.
func TestUserAdd(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		role        model.UserRole
		permError   bool
		force       bool
		dryRun      bool
		skipEmails  bool
		skipLMSSync bool
		newUsers    []*core.UserInfoWithPass
		expected    AddResponse
	}{
		// Note that the order on output user sync slices is not guarenteed.

		// Basic
		{
			model.RoleAdmin, false,
			false, false, true, false,
			[]*core.UserInfoWithPass{
				&core.UserInfoWithPass{core.UserInfo{"add@test.com", "add", model.RoleAdmin, "add@test.com"}, ""},
				&core.UserInfoWithPass{core.UserInfo{"student@test.com", "new name", model.RoleStudent, ""}, ""},
			},
			AddResponse{
				SyncUsersInfo: core.SyncUsersInfo{
					Add: []*core.UserInfo{
						&core.UserInfo{"add@test.com", "add", model.RoleAdmin, "lms-add@test.com"},
					},
					Mod: []*core.UserInfo{},
					Del: []*core.UserInfo{},
					Skip: []*core.UserInfo{
						&core.UserInfo{"student@test.com", "student", model.RoleStudent, "lms-student@test.com"},
					},
					Unchanged: []*core.UserInfo{},
				},
				Errors:       []AddError{},
				LMSSyncCount: 2,
			},
		},

		// User Errors
		{
			model.RoleAdmin, false,
			false, false, true, false,
			[]*core.UserInfoWithPass{
				&core.UserInfoWithPass{core.UserInfo{"", "", model.RoleStudent, ""}, ""},
				&core.UserInfoWithPass{core.UserInfo{"owner@test.com", "new name", model.RoleOwner, ""}, ""},
				&core.UserInfoWithPass{core.UserInfo{"owner@test.com", "", model.RoleAdmin, ""}, ""},
			},
			AddResponse{
				SyncUsersInfo: core.SyncUsersInfo{
					Add:       []*core.UserInfo{},
					Mod:       []*core.UserInfo{},
					Del:       []*core.UserInfo{},
					Skip:      []*core.UserInfo{},
					Unchanged: []*core.UserInfo{},
				},
				Errors: []AddError{
					AddError{0, "", "Empty emails are not allowed."},
					AddError{1, "owner@test.com", "Cannot create a user with a higher role (owner) than your role (admin)."},
					AddError{2, "owner@test.com", "Cannot modify a user with a higher role (owner) than your role (admin)."},
				},
				LMSSyncCount: 0,
			},
		},

		// Perm Error
		{
			model.RoleGrader, true,
			false, false, true, false,
			[]*core.UserInfoWithPass{},
			AddResponse{},
		},

		// No LMS
		{
			model.RoleAdmin, false,
			false, false, true, true,
			[]*core.UserInfoWithPass{
				&core.UserInfoWithPass{core.UserInfo{"add@test.com", "add", model.RoleAdmin, ""}, ""},
				&core.UserInfoWithPass{core.UserInfo{"student@test.com", "new name", model.RoleStudent, ""}, ""},
			},
			AddResponse{
				SyncUsersInfo: core.SyncUsersInfo{
					Add: []*core.UserInfo{
						&core.UserInfo{"add@test.com", "add", model.RoleAdmin, ""},
					},
					Mod: []*core.UserInfo{},
					Del: []*core.UserInfo{},
					Skip: []*core.UserInfo{
						&core.UserInfo{"student@test.com", "student", model.RoleStudent, ""},
					},
					Unchanged: []*core.UserInfo{},
				},
				Errors:       []AddError{},
				LMSSyncCount: 0,
			},
		},

		// Empty
		{
			model.RoleAdmin, false,
			false, false, true, false,
			[]*core.UserInfoWithPass{},
			AddResponse{
				SyncUsersInfo: core.SyncUsersInfo{
					Add:       []*core.UserInfo{},
					Mod:       []*core.UserInfo{},
					Del:       []*core.UserInfo{},
					Skip:      []*core.UserInfo{},
					Unchanged: []*core.UserInfo{},
				},
				Errors:       []AddError{},
				LMSSyncCount: 0,
			},
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		fields := map[string]any{
			"force":         testCase.force,
			"dry-run":       testCase.dryRun,
			"skip-emails":   testCase.skipEmails,
			"skip-lms-sync": testCase.skipLMSSync,
			"new-users":     testCase.newUsers,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`user/add`), fields, nil, testCase.role)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-020"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expcted '%s', found '%s'.",
						i, expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		var responseContent AddResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(testCase.expected, responseContent) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', actual: '%s'.", i,
				util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}
	}
}
