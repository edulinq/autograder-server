package submission

import (
    "reflect"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func TestHistory(test *testing.T) {
    testCases := []struct{ role usr.UserRole; target string; found bool; permError bool; expected []*artifact.SubmissionSummary; }{
        // Self.
        {usr.Student, "", true, false, studentHist},
        {usr.Grader, "", true, false, []*artifact.SubmissionSummary{}},

        // Other
        {usr.Grader, "student@test.com", true, false, studentHist},
        {usr.Student, "grader@test.com", true, true, nil},

        // Missing user.
        {usr.Student, "ZZZ@test.com", false, true, nil},
        {usr.Grader, "ZZZ@test.com", false, false, []*artifact.SubmissionSummary{}},
    };

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.target,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/history`), fields, nil, testCase.role);
        if (!response.Success) {
            if (testCase.permError) {
                expectedLocator := "-319";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }

            continue;
        }

        var responseContent HistoryResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (testCase.found != responseContent.FoundUser) {
            test.Errorf("Case %d: FoundUser value mismatch. Expected: '%v', actual: '%v'.", i, testCase.found, responseContent.FoundUser);
            continue;
        }

        if (responseContent.History == nil) {
            test.Errorf("Case %d: History is nil when is should not be: '%v'.", i, response);
            continue;
        }

        if (!reflect.DeepEqual(testCase.expected, responseContent.History)) {
            test.Errorf("Case %d: History does not match. Expected: '%+v', actual: '%+v'.", i, testCase.expected, responseContent.History);
            continue;
        }
    }
}

var studentHist []*artifact.SubmissionSummary = []*artifact.SubmissionSummary{
    &artifact.SubmissionSummary{
        ID: "course101::hw0::student@test.com::1697406256",
        Message: "",
        MaxPoints: 2,
        Score: 0,
        GradingStartTime: util.MustFromTimestamp("2023-10-15T21:44:16.84006Z"),
    },
    &artifact.SubmissionSummary{
        ID: "course101::hw0::student@test.com::1697406265",
        Message: "",
        MaxPoints: 2,
        Score: 1,
        GradingStartTime: util.MustFromTimestamp("2023-10-15T21:44:26.445382Z"),
    },
    &artifact.SubmissionSummary{
        ID: "course101::hw0::student@test.com::1697406272",
        Message: "",
        MaxPoints: 2,
        Score: 2,
        GradingStartTime: util.MustFromTimestamp("2023-10-15T21:44:33.157275Z"),
    },
};
