package user

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/usr"
)

func TestUserGet(test *testing.T) {
    testCases := []struct{ email string; expected *core.UserInfo }{
        {"other@test.com", &core.UserInfo{"other@test.com", "other", usr.Other, ""}},
        {"student@test.com", &core.UserInfo{"student@test.com", "student", usr.Student, ""}},
        {"grader@test.com", &core.UserInfo{"grader@test.com", "grader", usr.Grader, ""}},
        {"admin@test.com", &core.UserInfo{"admin@test.com", "admin", usr.Admin, ""}},
        {"owner@test.com", &core.UserInfo{"owner@test.com", "owner", usr.Owner, ""}},

        {"ZZZ", nil},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.email,
        };

        response := core.SendTestAPIRequest(test, core.NewEndpoint(`user/get`), fields);
        if (!response.Success) {
            test.Errorf("Case %d: Response is not a success: '%v'.", i, response);
            continue;
        }

        responseContent := response.Content.(map[string]any);

        expectedFound := (testCase.expected != nil);
        actualFound := responseContent["found"].(bool);
        if (expectedFound != actualFound) {
            test.Errorf("Case %d: Found user does not match. Expected: '%v', actual: '%v'.", i, expectedFound, actualFound);
            continue;
        }

        if (responseContent["user"] == nil) {
            if (testCase.expected != nil) {
                test.Errorf("Case %d: Got a nil user when one was expected.", i);
                continue;
            }

            continue;
        }

        actualUser := core.UserInfoFromMap(responseContent["user"].(map[string]any));
        if (*testCase.expected != *actualUser) {
            test.Errorf("Case %d: Unexpected user result. Expected: '%+v', actual: '%+v'.", i, testCase.expected, actualUser);
            continue;
        }
    }
}
