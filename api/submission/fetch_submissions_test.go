package submission

import (
    "reflect"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

var hash string = "45adfd6ba2624ec5b7b3ae4a0f11b9077b27db36837659b7f30ad13f1406ea34";

func TestFetchSubmissions(test *testing.T) {
    // See writeZipContents() for hash information.
    testCases := []struct{
            role usr.UserRole
            permError bool
            submissionIDs map[string]string
            hash string
    }{
        {usr.Other, true, nil, ""},
        {usr.Student, true, nil, ""},
        {usr.Grader, false, map[string]string{"student@test.com": "1697406272"}, hash},
        {usr.Admin, false, map[string]string{"student@test.com": "1697406272"}, hash},
        {usr.Owner, false, map[string]string{"student@test.com": "1697406272"}, hash},
    };

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    for i, testCase := range testCases {
        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/fetch/submissions`), nil, nil, testCase.role);
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

        var responseContent FetchSubmissionsResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (!reflect.DeepEqual(testCase.submissionIDs, responseContent.SubmissionIDs)) {
            test.Errorf("Case %d: Unexpected submission IDs. Expected: '%+v', actual: '%+v'.", i, testCase.submissionIDs, responseContent.SubmissionIDs);
            continue;
        }

        actualHash := util.Sha256HexFromString(responseContent.Contents);
        if (testCase.hash != actualHash) {
            test.Errorf("Case %d: Unexpected submissions hash. Expected: '%+v', actual: '%+v'.", i, testCase.hash, actualHash);
            continue;
        }
    }
}
