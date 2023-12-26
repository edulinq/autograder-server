package canvas

import (
    "testing"

    "github.com/eriq-augustine/autograder/lms/lmstypes"
    "github.com/eriq-augustine/autograder/model"
)

func TestCanvasUserGetBase(test *testing.T) {
    testCases := []struct{email string; expected *lmstypes.User}{
        {
            "owner@test.com",
            &lmstypes.User{
                ID: "00010",
                Name: "owner",
                Email: "owner@test.com",
                Role: model.RoleOwner,
            },
        },
        {
            "admin@test.com",
            &lmstypes.User{
                ID: "00020",
                Name: "admin",
                Email: "admin@test.com",
                Role: model.RoleAdmin,
            },
        },
        {
            "student@test.com",
            &lmstypes.User{
                ID: "00040",
                Name: "student",
                Email: "student@test.com",
                Role: model.RoleStudent,
            },
        },
    };

    for i, testCase := range testCases {
        user, err := testBackend.FetchUser(testCase.email);
        if (err != nil) {
            test.Errorf("Case %d: Failed to fetch user: '%v'.", i, err);
            continue;
        }

        if (*testCase.expected != *user) {
            test.Errorf("Case %d: User not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.expected, user);
            continue;
        }
    }
}
