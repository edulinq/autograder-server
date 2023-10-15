package api

import (
    "testing"

    "github.com/eriq-augustine/autograder/usr"
)

func TestUserGet(test *testing.T) {
    testCases := []struct{ email string; expected *userListRow }{
        {"other@test.com", &userListRow{"other@test.com", "other", usr.Other}},
        {"student@test.com", &userListRow{"student@test.com", "student", usr.Student}},
        {"grader@test.com", &userListRow{"grader@test.com", "grader", usr.Grader}},
        {"admin@test.com", &userListRow{"admin@test.com", "admin", usr.Admin}},
        {"owner@test.com", &userListRow{"owner@test.com", "owner", usr.Owner}},

        {"ZZZ", nil},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "email": testCase.email,
        };

        response := sendTestAPIRequest(test, "/api/v02/user/get", fields);
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

        actualUserInfo := responseContent["user"].(map[string]any);

        actualUser := &userListRow{
            Email: actualUserInfo["email"].(string),
            Name: actualUserInfo["name"].(string),
            Role: usr.GetRole(actualUserInfo["role"].(string)),
        };

        if (*testCase.expected != *actualUser) {
            test.Errorf("Case %d: Unexpected user result. Expected: '%v', actual: '%v'.", i, testCase.expected, actualUser);
            continue;
        }
    }
}
