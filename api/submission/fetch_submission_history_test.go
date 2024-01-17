package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
    "path/filepath"
    "reflect"
    "testing"
)

func TestFetchSubmissionHistory(test *testing.T) {
    // Note that computation of these paths is deferred until test time.
    studentGradingResults := []*model.GradingResult{
        model.MustLoadGradingResult(getTestSubmissionsResultPath("1697406256")),
        model.MustLoadGradingResult(getTestSubmissionsResultPath("1697406265")),
        model.MustLoadGradingResult(getTestSubmissionsResultPath("1697406272")),
    };

    testCases := []struct{
        role model.UserRole
        targetEmail string
        foundUser bool
        foundSubmissions bool
        permError bool
        result []*model.GradingResult
    }{
        // Grader, self.
        {model.RoleGrader, "",                 true, false, false, nil},
        {model.RoleGrader, "grader@test.com",  true, false, false, nil},

        // Grader, other.
        {model.RoleGrader, "student@test.com", true, true, false, studentGradingResults},

        // Grader, missing.
        {model.RoleGrader, "ZZZ@test.com",     false, false, false, nil},

        // Student, self.
        {model.RoleStudent, "",                true, true, true, studentGradingResults},

        // Student, self, missing.
        {model.RoleStudent, "",                true, false, true, nil},
        {model.RoleStudent, "student@test.com",true, false, true, nil},

        // Student, other, recent.
        {model.RoleStudent, "grader@test.com", false, false, true, nil},

        // Student, other, missing.
        {model.RoleStudent, "grader@test.com", true, false, true, nil},

        // Owner, self.
        {model.RoleOwner,   "",                true, false, false, nil},
        {model.RoleOwner,   "owner@test.com",  true, false, false, nil},

        // Owner, other.
        {model.RoleOwner,   "student@test.com",true, true, false, studentGradingResults},

        // Owner, missing.
        {model.RoleOwner,   "ZZZ@test.com",    false, false, false, nil},

        // Admin, self.
        {model.RoleAdmin,   "",                true, false, false, nil},
        {model.RoleAdmin,   "owner@test.com",  true, false, false, nil},

        // Admin, other.
        {model.RoleAdmin,   "student@test.com",true, true, false, studentGradingResults},

        // Admin, missing.
        {model.RoleAdmin,   "ZZZ@test.com",    false, false, false, nil},

    };

    for i, testCase := range testCases {
        field := map[string]any{
            "target-email": testCase.targetEmail,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/fetch/history`), field, nil, testCase.role);
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

        var responseContent FetchSubmissionHistoryResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (testCase.foundUser != responseContent.FoundUser) {
            test.Errorf("Case %d: Found user does not match. Expected: '%v', actual: '%v'.", i, testCase.foundUser, responseContent.FoundUser);
            continue;
        }

        if (testCase.foundSubmissions != responseContent.FoundSubmissions) {
            test.Errorf("Case %d: Found submissions does not match. Expected: '%v', actual: '%v'.", i, testCase.foundSubmissions, responseContent.FoundSubmissions);
            test.Errorf("Expected: %v, actual: %v", testCase.result, responseContent.GradingResults)
            continue;
        }

        if (!responseContent.FoundSubmissions) {
            continue;
        }

        if (!reflect.DeepEqual(testCase.result, responseContent.GradingResults)) {
            test.Errorf("Case %d: Unexpected submission result. Expected: '%s', actual: '%s'.", i,
                util.MustToJSONIndent(testCase.result), util.MustToJSONIndent(responseContent.GradingResults));
            continue;
        }
    }
}

func getTestSubmissionsResultPath(shortID string) string {
    return filepath.Join(config.GetCourseImportDir(), "_tests", "COURSE101", "submissions", "HW0", "student@test.com", shortID, "submission-result.json");
}
