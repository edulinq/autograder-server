package user

import (
    "slices"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/usr"
)

func TestList(test *testing.T) {
    expectedUsers := []core.UserInfo{
        core.UserInfo{"other@test.com", "other", usr.Other, ""},
        core.UserInfo{"student@test.com", "student", usr.Student, ""},
        core.UserInfo{"grader@test.com", "grader", usr.Grader, ""},
        core.UserInfo{"admin@test.com", "admin", usr.Admin, ""},
        core.UserInfo{"owner@test.com", "owner", usr.Owner, ""},
    };

    response := core.SendTestAPIRequest(test, core.NewEndpoint(`user/list`), nil);
    if (!response.Success) {
        test.Fatalf("Response is not a success: '%v'.", response);
    }

    responseContent := response.Content.(map[string]any);

    if (responseContent["users"] == nil) {
        test.Fatalf("Got a nil user list.");
    }

    rawUsers := responseContent["users"].([]any);
    actualUsers := make([]core.UserInfo, 0, len(rawUsers));

    for _, rawUser := range rawUsers {
        actualUsers = append(actualUsers, *core.UserInfoFromMap(rawUser.(map[string]any)));
    }

    slices.SortFunc(expectedUsers, core.CompareUserInfo);
    slices.SortFunc(actualUsers, core.CompareUserInfo);

    if (!slices.Equal(expectedUsers, actualUsers)) {
        test.Fatalf("Users not as expected. Expected: '%+v', actual: '%+v'.", expectedUsers, actualUsers);
    }
}
