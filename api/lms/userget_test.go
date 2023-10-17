package lms

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func TestUserGet(test *testing.T) {
    testCases := []struct{ email string; expected *core.UserInfo }{
        {"other@test.com", &core.UserInfo{"other@test.com", "other", usr.Other, "lms-other@test.com"}},
        {"student@test.com", &core.UserInfo{"student@test.com", "student", usr.Student, "lms-student@test.com"}},
        {"grader@test.com", &core.UserInfo{"grader@test.com", "grader", usr.Grader, "lms-grader@test.com"}},
        {"admin@test.com", &core.UserInfo{"admin@test.com", "admin", usr.Admin, "lms-admin@test.com"}},
        {"owner@test.com", &core.UserInfo{"owner@test.com", "owner", usr.Owner, "lms-owner@test.com"}},

        {"ZZZ", nil},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.email,
        };

        response := core.SendTestAPIRequest(test, core.NewEndpoint(`lms/user/get`), fields);
        if (!response.Success) {
            test.Errorf("Case %d: Response is not a success: '%v'.", i, response);
            continue;
        }

        var responseContent UserGetResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (!responseContent.FoundAGUser) {
            if (testCase.expected != nil) {
                test.Errorf("Case %d: Did not find a user when it was expected: '%v'.", i, testCase.expected);
            }

            continue;
        }

        if (responseContent.User == nil) {
            test.Errorf("Case %d: Got a nil user when one was expected.", i);
            continue;
        }

        if (*testCase.expected != *responseContent.User) {
            test.Errorf("Case %d: Unexpected user result. Expected: '%+v', actual: '%+v'.",
                    i, testCase.expected, responseContent.User);
            continue;
        }
    }
}
