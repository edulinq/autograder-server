package user

import (
    "reflect"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

func TestUserGet(test *testing.T) {
    testCases := []struct{ role model.UserRole; target string; permError bool; expected *core.UserInfo }{
        {model.RoleGrader, "other@test.com", false, &core.UserInfo{"other@test.com", "other", model.RoleOther, ""}},
        {model.RoleGrader, "student@test.com", false, &core.UserInfo{"student@test.com", "student", model.RoleStudent, ""}},
        {model.RoleGrader, "grader@test.com", false, &core.UserInfo{"grader@test.com", "grader", model.RoleGrader, ""}},
        {model.RoleGrader, "admin@test.com", false, &core.UserInfo{"admin@test.com", "admin", model.RoleAdmin, ""}},
        {model.RoleGrader, "owner@test.com", false, &core.UserInfo{"owner@test.com", "owner", model.RoleOwner, ""}},

        {model.RoleStudent, "student@test.com", true, nil},

        {model.RoleGrader, "", false, nil},

        {model.RoleGrader, "ZZZ", false, nil},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.target,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`user/get`), fields, nil, testCase.role);
        if (!response.Success) {
            expectedLocator := "";
            if (testCase.permError) {
                expectedLocator = "-020";
            } else if (testCase.target == "") {
                expectedLocator = "-034";
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
        if (expectedFound != responseContent.FoundUser) {
            test.Errorf("Case %d: Found user does not match. Expected: '%v', actual: '%v'.", i, expectedFound, responseContent.FoundUser);
            continue;
        }

        if (!reflect.DeepEqual(testCase.expected, responseContent.User)) {
            test.Errorf("Case %d: Unexpected user result. Expected: '%+v', actual: '%+v'.", i, testCase.expected, responseContent.User);
            continue;
        }
    }
}
