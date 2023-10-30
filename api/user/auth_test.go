package user

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func TestAuth(test *testing.T) {
    testCases := []struct{ role usr.UserRole; email string; pass string; expected AuthResponse }{
        {usr.Grader, "other@test.com",   "other",   AuthResponse{true, true}},
        {usr.Grader, "student@test.com", "student", AuthResponse{true, true}},
        {usr.Grader, "grader@test.com",  "grader",  AuthResponse{true, true}},
        {usr.Grader, "admin@test.com",   "admin",   AuthResponse{true, true}},
        {usr.Grader, "owner@test.com",   "owner",   AuthResponse{true, true}},

        {usr.Grader, "other@test.com",   "ZZZ", AuthResponse{true, false}},
        {usr.Grader, "student@test.com", "ZZZ", AuthResponse{true, false}},
        {usr.Grader, "grader@test.com",  "ZZZ", AuthResponse{true, false}},
        {usr.Grader, "admin@test.com",   "ZZZ", AuthResponse{true, false}},
        {usr.Grader, "owner@test.com",   "ZZZ", AuthResponse{true, false}},

        {usr.Other,   "student@test.com", "student", AuthResponse{true, true}},
        {usr.Student, "student@test.com", "student", AuthResponse{true, true}},
        {usr.Grader,  "student@test.com", "student", AuthResponse{true, true}},
        {usr.Admin,   "student@test.com", "student", AuthResponse{true, true}},
        {usr.Owner,   "student@test.com", "student", AuthResponse{true, true}},

        {usr.Grader, "ZZZ", "ZZZ", AuthResponse{false, false}},
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
