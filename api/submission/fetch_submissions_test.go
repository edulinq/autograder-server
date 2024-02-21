package submission

import (
    "reflect"
    "testing"

    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/util"
)

func TestFetchSubmissions(test *testing.T) {
    testCases := []struct{
            role model.UserRole
            permError bool
    }{
        {model.RoleOther, true},
        {model.RoleStudent, true},
        {model.RoleGrader, false},
        {model.RoleAdmin, false},
        {model.RoleOwner, false},
    };

    submissions := map[string]*model.GradingResult{
        "other@test.com": nil,
        "student@test.com": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
        "grader@test.com": nil,
        "admin@test.com": nil,
        "owner@test.com": nil,
    };

    for i, testCase := range testCases {
        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/fetch/submissions`), nil, nil, testCase.role);
        if (!response.Success) {
            if (testCase.permError) {
                expectedLocator := "-020";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }

            continue;
        }

        var responseContent FetchSubmissionsResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (!reflect.DeepEqual(submissions, responseContent.GradingResults)) {
            test.Errorf("Case %d: Unexpected submission IDs. Expected: '%s', actual: '%s'.", i,
                    util.MustToJSONIndent(submissions), util.MustToJSONIndent(responseContent.GradingResults));
            continue;
        }
    }
}
