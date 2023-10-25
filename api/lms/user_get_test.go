package lms

import (
    "reflect"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func TestUserGet(test *testing.T) {
    testCases := []struct{ role usr.UserRole; target string; permError bool; expected *core.UserInfo }{
        {usr.Grader, "other@test.com", false, &core.UserInfo{"other@test.com", "other", usr.Other, "lms-other@test.com"}},
        {usr.Grader, "student@test.com", false, &core.UserInfo{"student@test.com", "student", usr.Student, "lms-student@test.com"}},
        {usr.Grader, "grader@test.com", false, &core.UserInfo{"grader@test.com", "grader", usr.Grader, "lms-grader@test.com"}},
        {usr.Grader, "admin@test.com", false, &core.UserInfo{"admin@test.com", "admin", usr.Admin, "lms-admin@test.com"}},
        {usr.Grader, "owner@test.com", false, &core.UserInfo{"owner@test.com", "owner", usr.Owner, "lms-owner@test.com"}},

        {usr.Student, "student@test.com", true, nil},

        {usr.Grader, "", false, nil},

        {usr.Grader, "ZZZ", false, nil},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.target,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`lms/user/get`), fields, nil, testCase.role);
        if (!response.Success) {
            expectedLocator := "";
            if (testCase.permError) {
                expectedLocator = "-306";
            } else if (testCase.target == "") {
                expectedLocator = "-320";
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

        var responseContent UserGetResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        expectedFound := (testCase.expected != nil);
        if (expectedFound != responseContent.FoundAGUser) {
            test.Errorf("Case %d: Found AG user does not match. Expected: '%v', actual: '%v'.", i, expectedFound, responseContent.FoundAGUser);
            continue;
        }

        if (expectedFound != responseContent.FoundLMSUser) {
            test.Errorf("Case %d: Found LMS user does not match. Expected: '%v', actual: '%v'.", i, expectedFound, responseContent.FoundLMSUser);
            continue;
        }

        if (!reflect.DeepEqual(testCase.expected, responseContent.User)) {
            test.Errorf("Case %d: Unexpected user result. Expected: '%+v', actual: '%+v'.", i, testCase.expected, responseContent.User);
            continue;
        }
    }
}
