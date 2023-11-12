package user

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func TestRemove(test *testing.T) {
    // Leave the course in a good state after the test.
    defer db.ResetForTesting();

    testCases := []struct{ role usr.UserRole; target string; basicPermError bool; advPermError bool; expected RemoveResponse }{
        {usr.Owner, "other@test.com",   false, false, RemoveResponse{true}},
        {usr.Owner, "student@test.com", false, false, RemoveResponse{true}},
        {usr.Owner, "grader@test.com",  false, false, RemoveResponse{true}},
        {usr.Owner, "admin@test.com",   false, false, RemoveResponse{true}},
        {usr.Owner, "owner@test.com",   false, false, RemoveResponse{true}},

        {usr.Other,   "other@test.com", true,  false, RemoveResponse{true}},
        {usr.Student, "other@test.com", true,  false, RemoveResponse{true}},
        {usr.Grader,  "other@test.com", true,  false, RemoveResponse{true}},
        {usr.Admin,   "other@test.com", false, false, RemoveResponse{true}},
        {usr.Owner,   "other@test.com", false, false, RemoveResponse{true}},

        {usr.Owner, "ZZZ", false, false, RemoveResponse{false}},

        {usr.Admin, "owner@test.com", false, true,  RemoveResponse{true}},
        {usr.Owner, "owner@test.com", false, false, RemoveResponse{true}},
    };

    for i, testCase := range testCases {
        // Reload the test course every time.
        db.ResetForTesting();

        fields := map[string]any{
            "target-email": testCase.target,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`user/remove`), fields, nil, testCase.role);

        if (!response.Success) {
            expectedLocator := "";
            if (testCase.basicPermError) {
                expectedLocator = "-306";
            } else if (testCase.advPermError) {
                expectedLocator = "-601";
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

        var responseContent RemoveResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (testCase.expected != responseContent) {
            test.Errorf("Case %d: Unexpected result. Expected: '%+v', actual: '%+v'.", i, testCase.expected, responseContent);
            continue;
        }

        if (!testCase.expected.FoundUser) {
            continue;
        }

        // Ensure that the user is removed.

        course := db.MustGetCourse("course101");
        user, err := db.GetUser(course, testCase.target);
        if (err != nil) {
            test.Errorf("Case %d: Failed to get removed user: '%v'.", i, err);
            continue;
        }

        if (user != nil) {
            test.Errorf("Case %d: User not removed from DB: '%+v'.", i, user);
            continue;
        }
    }
}
