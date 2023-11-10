package submission

import (
    "maps"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func TestFetchScores(test *testing.T) {
    testCases := []struct{ role usr.UserRole; filterRole usr.UserRole; permError bool; ids map[string]string; }{
        {usr.Grader, usr.Unknown, false, map[string]string{
            "other@test.com": "",
            "student@test.com": "course101::hw0::student@test.com::1697406272",
            "grader@test.com": "",
            "admin@test.com": "",
            "owner@test.com": "",
        }},
        {usr.Admin, usr.Unknown, false, map[string]string{
            "other@test.com": "",
            "student@test.com": "course101::hw0::student@test.com::1697406272",
            "grader@test.com": "",
            "admin@test.com": "",
            "owner@test.com": "",
        }},
        {usr.Grader, usr.Student, false, map[string]string{
            "student@test.com": "course101::hw0::student@test.com::1697406272",
        }},
        {usr.Grader, usr.Grader, false, map[string]string{
            "grader@test.com": "",
        }},
        {usr.Student, usr.Unknown, true, nil},
        {usr.Student, usr.Student, true, nil},
        {usr.Other, usr.Unknown, true, nil},
        {usr.Other, usr.Grader, true, nil},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "filter-role": testCase.filterRole,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/fetch/scores`), fields, nil, testCase.role);
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

        var responseContent FetchScoresResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        actualIDs := make(map[string]string, len(testCase.ids));
        for email, summary := range responseContent.SubmissionInfos {
            id := "";
            if (summary != nil) {
                id = summary.ID;
            }

            actualIDs[email] = id;
        }

        if (!maps.Equal(testCase.ids, actualIDs)) {
            test.Errorf("Case %d: Submission IDs do not match. Expected: '%+v', actual: '%+v'.", i, testCase.ids, actualIDs);
            continue;
        }
    }
}
