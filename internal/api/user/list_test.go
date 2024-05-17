package user

import (
    "slices"
    "testing"

    "github.com/edulinq/autograder/internal/api/core"
    "github.com/edulinq/autograder/internal/model"
)

func TestUserList(test *testing.T) {
    expectedUsers := []core.UserInfo{
        core.UserInfo{"other@test.com", "other", model.RoleOther, ""},
        core.UserInfo{"student@test.com", "student", model.RoleStudent, ""},
        core.UserInfo{"grader@test.com", "grader", model.RoleGrader, ""},
        core.UserInfo{"admin@test.com", "admin", model.RoleAdmin, ""},
        core.UserInfo{"owner@test.com", "owner", model.RoleOwner, ""},
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
