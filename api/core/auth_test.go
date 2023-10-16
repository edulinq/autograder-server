package core

import (
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

func TestAuth(test *testing.T) {
    type baseAPIRequest struct {
        APIRequestCourseUserContext
        MinRoleOther
    }

    testCases := []struct{email string; pass string; noauth bool; locator string}{
        {"owner@test.com",   "owner",    false, ""},
        {"admin@test.com",   "admin",    false, ""},
        {"grader@test.com",  "grader",   false, ""},
        {"student@test.com", "student",  false, ""},
        {"other@test.com",   "other",    false, ""},

        {"Z",                 "student",  false, "-202"},
        {"Zstudent@test.com", "student",  false, "-202"},
        {"student@test.comZ", "student",  false, "-202"},
        {"student",           "student",  false, "-202"},

        {"student@test.com", "",         false, "-203"},
        {"student@test.com", "Zstudent", false, "-203"},
        {"student@test.com", "studentZ", false, "-203"},

        {"owner@test.com",   "owner",    true, ""},
        {"admin@test.com",   "admin",    true, ""},
        {"grader@test.com",  "grader",   true, ""},
        {"student@test.com", "student",  true, ""},
        {"other@test.com",   "other",    true, ""},

        {"Z",                 "student",  true, "-202"},
        {"Zstudent@test.com", "student",  true, "-202"},
        {"student@test.comZ", "student",  true, "-202"},
        {"student",           "student",  true, "-202"},

        {"student@test.com", "",         true, ""},
        {"student@test.com", "Zstudent", true, ""},
        {"student@test.com", "studentZ", true, ""},
    };

    oldNoAuth := config.NO_AUTH.GetBool();
    defer config.NO_AUTH.Set(oldNoAuth);

    for i, testCase := range testCases {
        request := baseAPIRequest{
            APIRequestCourseUserContext: APIRequestCourseUserContext{
                CourseID: "COURSE101",
                UserEmail: testCase.email,
                UserPass: util.Sha256HexFromString(testCase.pass),
            },
        };

        config.NO_AUTH.Set(testCase.noauth);
        apiErr := ValidateAPIRequest(nil, &request, "");

        if ((apiErr == nil) && (testCase.locator != "")) {
            test.Errorf("Case %d: Expecting error '%s', but got no error.", i, testCase.locator);
        } else if ((apiErr != nil) && (testCase.locator == "")) {
            test.Errorf("Case %d: Expecting no error, but got '%s': '%v'.", i, apiErr.Locator, apiErr);
        } else if ((apiErr != nil) && (testCase.locator != "") && (apiErr.Locator != testCase.locator)) {
            test.Errorf("Case %d: Got a different error than expected. Expected: '%s', actual: '%s' -- '%v'.",
                    i, testCase.locator, apiErr.Locator, apiErr);
        }
    }
}
