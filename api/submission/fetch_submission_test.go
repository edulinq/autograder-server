package submission

import (
    "path/filepath"
    "reflect"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

func TestFetchSubmission(test *testing.T) {
    // Note that computation of these paths is deferred until test time.
    studentGradingResults := map[string]*model.GradingResult{
        "1697406256": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406256")),
        "1697406265": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406265")),
        "1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
    };

    testCases := []struct{
            role model.UserRole
            targetEmail string
            targetSubmission string
            foundUser bool
            foundSubmission bool
            permError bool
            result *model.GradingResult
    }{
        // Grader, self, recent.
        {model.RoleGrader, "",                "", true, false, false, nil},
        {model.RoleGrader, "grader@test.com", "", true, false, false, nil},

        // Grader, self, missing.
        {model.RoleGrader, "",                "ZZZ", true, false, false, nil},
        {model.RoleGrader, "grader@test.com", "ZZZ", true, false, false, nil},

        // Grader, other, recent.
        {model.RoleGrader, "student@test.com", "", true, true, false, studentGradingResults["1697406272"]},

        // Grader, other, specific.
        {model.RoleGrader, "student@test.com", "1697406256", true, true, false, studentGradingResults["1697406256"]},
        {model.RoleGrader, "student@test.com", "1697406265", true, true, false, studentGradingResults["1697406265"]},
        {model.RoleGrader, "student@test.com", "1697406272", true, true, false, studentGradingResults["1697406272"]},

        // Grader, other, specific (full ID).
        {model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406256", true, true, false, studentGradingResults["1697406256"]},
        {model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406265", true, true, false, studentGradingResults["1697406265"]},
        {model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406272", true, true, false, studentGradingResults["1697406272"]},

        // Grader, other, missing.
        {model.RoleGrader, "student@test.com", "ZZZ", true, false, false, nil},

        // Grader, missing, recent.
        {model.RoleGrader, "ZZZ@test.com", "", false, false, false, nil},

        // Student, self, recent.
        {model.RoleStudent, "",                 "", true, true, false, studentGradingResults["1697406272"]},
        {model.RoleStudent, "student@test.com", "", true, true, false, studentGradingResults["1697406272"]},

        // Student, self, missing.
        {model.RoleStudent, "",                 "ZZZ", true, false, false, nil},
        {model.RoleStudent, "student@test.com", "ZZZ", true, false, false, nil},

        // Student, other, recent.
        {model.RoleStudent, "grader@test.com", "", false, false, true, nil},

        // Student, other, missing.
        {model.RoleStudent, "grader@test.com", "ZZZ", true, false, true, nil},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.targetEmail,
            "target-submission": testCase.targetSubmission,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/fetch/submission`), fields, nil, testCase.role);
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

        var responseContent FetchSubmissionResponse;
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

        if (!reflect.DeepEqual(testCase.result, responseContent.GradingResult)) {
            test.Errorf("Case %d: Unexpected submission result. Expected: '%s', actual: '%s'.", i,
                    util.MustToJSONIndent(testCase.result), util.MustToJSONIndent(responseContent.GradingResult));
            continue;
        }
    }
}

func getTestSubmissionResultPath(shortID string) string {
    return filepath.Join(config.GetCourseImportDir(), "_tests", "COURSE101", "submissions", "HW0", "student@test.com", shortID, "submission-result.json");
}
