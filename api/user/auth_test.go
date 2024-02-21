package user

import (
    "testing"

    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/util"
)

func TestUserAuth(test *testing.T) {
    testCases := []struct{ role model.UserRole; email string; pass string; expected AuthResponse }{
        {model.RoleGrader, "other@test.com",   "other",   AuthResponse{true, true}},
        {model.RoleGrader, "student@test.com", "student", AuthResponse{true, true}},
        {model.RoleGrader, "grader@test.com",  "grader",  AuthResponse{true, true}},
        {model.RoleGrader, "admin@test.com",   "admin",   AuthResponse{true, true}},
        {model.RoleGrader, "owner@test.com",   "owner",   AuthResponse{true, true}},

        {model.RoleGrader, "other@test.com",   "ZZZ", AuthResponse{true, false}},
        {model.RoleGrader, "student@test.com", "ZZZ", AuthResponse{true, false}},
        {model.RoleGrader, "grader@test.com",  "ZZZ", AuthResponse{true, false}},
        {model.RoleGrader, "admin@test.com",   "ZZZ", AuthResponse{true, false}},
        {model.RoleGrader, "owner@test.com",   "ZZZ", AuthResponse{true, false}},

        {model.RoleOther,   "student@test.com", "student", AuthResponse{true, true}},
        {model.RoleStudent, "student@test.com", "student", AuthResponse{true, true}},
        {model.RoleGrader,  "student@test.com", "student", AuthResponse{true, true}},
        {model.RoleAdmin,   "student@test.com", "student", AuthResponse{true, true}},
        {model.RoleOwner,   "student@test.com", "student", AuthResponse{true, true}},

        {model.RoleGrader, "ZZZ", "ZZZ", AuthResponse{false, false}},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.email,
            "target-pass": util.Sha256HexFromString(testCase.pass),
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`user/auth`), fields, nil, testCase.role);
        if (!response.Success) {
            test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            continue;
        }

        var responseContent AuthResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (testCase.expected != responseContent) {
            test.Errorf("Case %d: Unexpected result. Expected: '%+v', actual: '%+v'.", i, testCase.expected, responseContent);
            continue;
        }
    }
}
