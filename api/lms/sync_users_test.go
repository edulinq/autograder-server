package lms

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// lmsusers.SyncLMSUsers() is already heavily tested, focus on API-specific functionality.
func TestAPISyncUsers(test *testing.T) {
    testCases := []struct{ role model.UserRole; dryRun bool; skipEmails bool; permError bool }{
        {model.RoleOther, false, true, true},
        {model.RoleStudent, false, true, true},
        {model.RoleGrader, false, true, true},
        {model.RoleAdmin, false, true, false},
        {model.RoleOwner, false, true, false},

        {model.RoleAdmin, false, true, false},
        {model.RoleAdmin, false, false, false},
        {model.RoleAdmin, true, true, false},
        {model.RoleAdmin, true, false, false},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "dry-run": testCase.dryRun,
            "skip-emails": testCase.skipEmails,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`lms/sync/users`), fields, nil, testCase.role);
        if (!response.Success) {
            if (testCase.permError) {
                expectedLocator := "-306";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }

            continue;
        }

        // Ensure the response can unmarshal.
        var responseContent SyncUsersResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);
    }
}
