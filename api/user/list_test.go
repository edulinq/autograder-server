package user

import (
    "slices"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/usr"
)

func TestUserList(test *testing.T) {
    expectedUsers := []UserInfo{
        UserInfo{"other@test.com", "other", usr.Other, ""},
        UserInfo{"student@test.com", "student", usr.Student, ""},
        UserInfo{"grader@test.com", "grader", usr.Grader, ""},
        UserInfo{"admin@test.com", "admin", usr.Admin, ""},
        UserInfo{"owner@test.com", "owner", usr.Owner, ""},
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
    actualUsers := make([]UserInfo, 0, len(rawUsers));

    for _, rawUser := range rawUsers {
        actualUsers = append(actualUsers, *UserInfoFromMap(rawUser.(map[string]any)));
    }

    slices.SortFunc(expectedUsers, CompareUserInfo);
    slices.SortFunc(actualUsers, CompareUserInfo);

    if (!slices.Equal(expectedUsers, actualUsers)) {
        test.Fatalf("Users not as expected. Expected: '%+v', actual: '%+v'.", expectedUsers, actualUsers);
    }
}
