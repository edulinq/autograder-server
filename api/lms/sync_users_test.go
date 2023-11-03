package lms

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

// lmsusers.SyncLMSUsers() is already heavily tested, focus on API-specific functionality.
func TestAPISyncUsers(test *testing.T) {
    testCases := []struct{ role usr.UserRole; dryRun bool; sendEmails bool; permError bool }{
        {usr.Other, false, false, true},
        {usr.Student, false, false, true},
        {usr.Grader, false, false, true},
        {usr.Admin, false, false, false},
        {usr.Owner, false, false, false},

        {usr.Admin, false, false, false},
        {usr.Admin, false, true, false},
        {usr.Admin, true, false, false},
        {usr.Admin, true, true, false},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "dry-run": testCase.dryRun,
            "send-emails": testCase.sendEmails,
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
