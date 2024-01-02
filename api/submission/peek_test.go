package submission

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

func TestPeek(test *testing.T) {
    // There are two options, which makes for four general test cases.
    testCases := []struct{ role model.UserRole; targetEmail string; targetSubmission string; score float64; foundUser bool; foundSubmission bool; permError bool }{
        // Grader, self, recent.
        {model.RoleGrader, "",                "", -1.0, true, false, false},
        {model.RoleGrader, "grader@test.com", "", -1.0, true, false, false},

        // Grader, self, missing.
        {model.RoleGrader, "",                "ZZZ", -1.0, true, false, false},
        {model.RoleGrader, "grader@test.com", "ZZZ", -1.0, true, false, false},

        // Grader, other, recent.
        {model.RoleGrader, "student@test.com", "", 2.0, true, true, false},

        // Grader, other, specific.
        {model.RoleGrader, "student@test.com", "1697406256", 0.0, true, true, false},
        {model.RoleGrader, "student@test.com", "1697406265", 1.0, true, true, false},
        {model.RoleGrader, "student@test.com", "1697406272", 2.0, true, true, false},

        // Grader, other, specific (full ID).
        {model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406256", 0.0, true, true, false},
        {model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406265", 1.0, true, true, false},
        {model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406272", 2.0, true, true, false},

        // Grader, other, missing.
        {model.RoleGrader, "student@test.com", "ZZZ", -1.0, true, false, false},

        // Grader, missing, recent.
        {model.RoleGrader, "ZZZ@test.com", "", -1.0, false, false, false},

        // Student, self, recent.
        {model.RoleStudent, "",                 "", 2.0, true, true, false},
        {model.RoleStudent, "student@test.com", "", 2.0, true, true, false},

        // Student, self, missing.
        {model.RoleStudent, "",                 "ZZZ", -1.0, true, false, false},
        {model.RoleStudent, "student@test.com", "ZZZ", -1.0, true, false, false},

        // Student, other, recent.
        {model.RoleStudent, "grader@test.com", "", -1.0, false, false, true},

        // Student, other, missing.
        {model.RoleStudent, "grader@test.com", "ZZZ", -1.0, false, false, true},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.targetEmail,
            "target-submission": testCase.targetSubmission,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/peek`), fields, nil, testCase.role);
        if (!response.Success) {
            if (testCase.permError) {
                expectedLocator := "-033";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }

            continue;
        }

        var responseContent PeekResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (testCase.foundUser != responseContent.FoundUser) {
            test.Errorf("Case %d: Found user does not match. Expected: '%v', actual: '%v'.", i, testCase.foundUser, responseContent.FoundUser);
            continue;
        }

        if (testCase.foundSubmission != responseContent.FoundSubmission) {
            test.Errorf("Case %d: Found submission does not match. Expected: '%v', actual: '%v'.", i, testCase.foundSubmission, responseContent.FoundSubmission);
            continue;
        }

        if (!responseContent.FoundSubmission) {
            continue;
        }

        actualScore := responseContent.GradingInfo.Score;
        if (!util.IsClose(testCase.score, actualScore)) {
            test.Errorf("Case %d: Unexpected submission score. Expected: '%+v', actual: '%+v'.", i, testCase.score, actualScore);
            continue;
        }
    }
}
