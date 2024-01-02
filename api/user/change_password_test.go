package user

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// Many of the semantics of add users are tested at the course level,
// focus on the API here.
func TestChangePassword(test *testing.T) {
    defer db.ResetForTesting();

    testCases := []struct{
            role model.UserRole; permError bool; advPermError bool;
            target string; newPass string;
            foundUser bool; hasEmail bool;
    }{
        // Self (context)
        {model.RoleOther,   false, false, "", "new-pass", true, false},
        {model.RoleStudent, false, false, "", "new-pass", true, false},
        {model.RoleGrader,  false, false, "", "new-pass", true, false},
        {model.RoleAdmin,   false, false, "", "new-pass", true, false},
        {model.RoleOwner,   false, false, "", "new-pass", true, false},

        // Self (direct)
        {model.RoleOther,   false, false, "other@test.com",   "new-pass", true, false},
        {model.RoleStudent, false, false, "student@test.com", "new-pass", true, false},
        {model.RoleGrader,  false, false, "grader@test.com",  "new-pass", true, false},
        {model.RoleAdmin,   false, false, "admin@test.com",   "new-pass", true, false},
        {model.RoleOwner,   false, false, "owner@test.com",   "new-pass", true, false},

        // Other
        {model.RoleOther,   true,  false, "student@test.com", "new-pass", true, false},
        {model.RoleStudent, true,  false, "other@test.com",   "new-pass", true, false},
        {model.RoleGrader,  true,  false, "other@test.com",   "new-pass", true, false},
        {model.RoleAdmin,   false, false, "other@test.com",   "new-pass", true, false},
        {model.RoleOwner,   false, false, "other@test.com",   "new-pass", true, false},

        // Advanced Perm Error
        {model.RoleAdmin, false, true,  "owner@test.com", "new-pass", true, false},
        {model.RoleOwner, false, false, "admin@test.com", "new-pass", true, false},

        // Missing
        {model.RoleOther,   true,  false, "ZZZ@test.com", "new-pass", false, false},
        {model.RoleStudent, true,  false, "ZZZ@test.com", "new-pass", false, false},
        {model.RoleGrader,  true,  false, "ZZZ@test.com", "new-pass", false, false},
        {model.RoleAdmin,   false, false, "ZZZ@test.com", "new-pass", false, false},
        {model.RoleOwner,   false, false, "ZZZ@test.com", "new-pass", false, false},

        // Email
        {model.RoleAdmin,   false, false, "other@test.com", "", true, true},
        {model.RoleAdmin,   false, false, "admin@test.com", "", true, true},
    };

    for i, testCase := range testCases {
        db.ResetForTesting();
        email.ClearTestMessages();

        fields := map[string]any{
            "target-email": testCase.target,
            "new-pass": testCase.newPass,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`user/change/pass`), fields, nil, testCase.role);
        if (!response.Success) {
            expectedLocator := "";
            if (testCase.permError) {
                expectedLocator = "-033";
            } else if (testCase.advPermError) {
                expectedLocator = "-805";
            }

            if (expectedLocator == "") {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            } else {
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            }

            continue;
        }

        var responseContent ChangePasswordResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (testCase.foundUser != responseContent.FoundUser) {
            test.Errorf("Case %d: Unexpected found user. Expected: '%v', actual: '%v'.", i,
                    testCase.foundUser, responseContent.FoundUser);
            continue;
        }

        hasEmail := (len(email.GetTestMessages()) > 0);
        if (testCase.hasEmail != hasEmail) {
            test.Errorf("Case %d: Unexpected has email. Expected: '%v', actual: '%v'.", i,
                    testCase.hasEmail, hasEmail);
            continue;
        }
    }
}
