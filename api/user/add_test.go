package user

import (
    "reflect"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

// Many of the semantics of add users are tested at the course level,
// focus on the API here.
func TestAdd(test *testing.T) {
    defer db.ResetForTesting();

    testCases := []struct{
            role usr.UserRole; permError bool
            force bool; dryRun bool; sendEmails bool; skipLMSSync bool
            newUsers []*core.UserInfoWithPass
            expected AddResponse
    }{
        // Note that the order on output user sync slices is not guarenteed.

        // Basic
        {
            usr.Admin, false,
            false, false, false, false,
            []*core.UserInfoWithPass{
                &core.UserInfoWithPass{core.UserInfo{"add@test.com", "add", usr.Admin, ""}, ""},
                &core.UserInfoWithPass{core.UserInfo{"student@test.com", "new name", usr.Student, ""}, ""},
            },
            AddResponse{
                SyncUsersInfo: core.SyncUsersInfo{
                    Add: []*core.UserInfo{
                        &core.UserInfo{"add@test.com", "add", usr.Admin, ""},
                    },
                    Mod: []*core.UserInfo{},
                    Del: []*core.UserInfo{},
                    Skip: []*core.UserInfo{
                        &core.UserInfo{"student@test.com", "student", usr.Student, ""},
                    },
                },
                Errors: []AddError{},
                LMSSyncCount: 6,
            },
        },

        // User Errors
        {
            usr.Admin, false,
            false, false, false, false,
            []*core.UserInfoWithPass{
                &core.UserInfoWithPass{core.UserInfo{"", "", usr.Student, ""}, ""},
                &core.UserInfoWithPass{core.UserInfo{"owner@test.com", "new name", usr.Owner, ""}, ""},
            },
            AddResponse{
                SyncUsersInfo: core.SyncUsersInfo{
                    Add: []*core.UserInfo{},
                    Mod: []*core.UserInfo{},
                    Del: []*core.UserInfo{},
                    Skip: []*core.UserInfo{},
                },
                Errors: []AddError{
                    AddError{0, "", "Empty emails are not allowed."},
                    AddError{1, "owner@test.com", "Cannot create a user with a higher role (owner) than your role (admin)."},
                },
                LMSSyncCount: 5,
            },
        },

        // Perm Error
        {
            usr.Grader, true,
            false, false, false, false,
            []*core.UserInfoWithPass{},
            AddResponse{},
        },

        // No LMS
        {
            usr.Admin, false,
            false, false, false, true,
            []*core.UserInfoWithPass{
                &core.UserInfoWithPass{core.UserInfo{"add@test.com", "add", usr.Admin, ""}, ""},
                &core.UserInfoWithPass{core.UserInfo{"student@test.com", "new name", usr.Student, ""}, ""},
            },
            AddResponse{
                SyncUsersInfo: core.SyncUsersInfo{
                    Add: []*core.UserInfo{
                        &core.UserInfo{"add@test.com", "add", usr.Admin, ""},
                    },
                    Mod: []*core.UserInfo{},
                    Del: []*core.UserInfo{},
                    Skip: []*core.UserInfo{
                        &core.UserInfo{"student@test.com", "student", usr.Student, ""},
                    },
                },
                Errors: []AddError{},
                LMSSyncCount: 0,
            },
        },

        // Empty
        {
            usr.Admin, false,
            false, false, false, false,
            []*core.UserInfoWithPass{},
            AddResponse{
                SyncUsersInfo: core.SyncUsersInfo{
                    Add: []*core.UserInfo{},
                    Mod: []*core.UserInfo{},
                    Del: []*core.UserInfo{},
                    Skip: []*core.UserInfo{},
                },
                Errors: []AddError{},
                LMSSyncCount: 5,
            },
        },
    };

    for i, testCase := range testCases {
        db.ResetForTesting();

        fields := map[string]any{
            "force": testCase.force,
            "dry-run": testCase.dryRun,
            "send-emails": testCase.sendEmails,
            "skip-lms-sync": testCase.skipLMSSync,
            "new-users": testCase.newUsers,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`user/add`), fields, nil, testCase.role);
        if (!response.Success) {
            if (testCase.permError) {
                expectedLocator := "-306";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }

            continue;
        }

        var responseContent AddResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (!reflect.DeepEqual(testCase.expected, responseContent)) {
            test.Errorf("Case %d: Unexpected result. Expected: '%s', actual: '%s'.", i,
                    util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent));
            continue;
        }
    }
}
