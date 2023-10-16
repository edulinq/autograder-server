package submission

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func TestPeek(test *testing.T) {
    // There are two options, which makes for four general test cases.
    testCases := []struct{ role usr.UserRole; target string; id string; score float64; permissionsError bool }{
        // Grader, self, recent.
        {usr.Grader, "",                "", -1.0, false},
        {usr.Grader, "grader@test.com", "", -1.0, false},

        // Grader, self, missing.
        {usr.Grader, "",                "ZZZ", -1.0, false},
        {usr.Grader, "grader@test.com", "ZZZ", -1.0, false},

        // Grader, other, recent.
        {usr.Grader, "student@test.com", "", 2.0, false},

        // Grader, other, specific.
        {usr.Grader, "student@test.com", "1697406256", 0.0, false},
        {usr.Grader, "student@test.com", "1697406265", 1.0, false},
        {usr.Grader, "student@test.com", "1697406272", 2.0, false},

        // Grader, other, missing.
        {usr.Grader, "student@test.com", "ZZZ", -1.0, false},

        // Student, self, recent.
        {usr.Student, "",                 "", 2.0, false},
        {usr.Student, "student@test.com", "", 2.0, false},

        // Student, self, missing.
        {usr.Student, "",                 "ZZZ", -1.0, false},
        {usr.Student, "student@test.com", "ZZZ", -1.0, false},

        // Student, other, recent.
        {usr.Student, "grader@test.com", "", -1.0, true},

        // Student, other, missing.
        {usr.Student, "grader@test.com", "ZZZ", -1.0, true},
    };

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.target,
            "submission-id": testCase.id,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/peek`), fields, nil, testCase.role);
        if (!response.Success) {
            if (testCase.permissionsError) {
                expectedLocator := "-401";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }

            continue;
        }

        responseContent := response.Content.(map[string]any);

        expectedFound := (testCase.score >= 0.0);
        actualFound := responseContent["found"].(bool);
        if (expectedFound != actualFound) {
            test.Errorf("Case %d: Found does not match. Expected: '%v', actual: '%v'.", i, expectedFound, actualFound);
            continue;
        }

        if (responseContent["assignment"] == nil) {
            if (expectedFound) {
                test.Errorf("Case %d: Got a nil assignment when one was expected.", i);
                continue;
            }

            continue;
        }

        var assignment artifact.GradedAssignment;
        util.MustJSONFromString(util.MustToJSON(responseContent["assignment"]), &assignment);

        actualScore := assignment.Score();
        if (!util.IsClose(testCase.score, actualScore)) {
            test.Errorf("Case %d: Unexpected assignment score. Expected: '%+v', actual: '%+v'.", i, testCase.score, actualScore);
            continue;
        }
    }
}
