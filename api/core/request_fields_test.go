package core

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "reflect"
    "strings"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

// Test CourseUsers and TargetUserSelfOrGrader.
// No embeded course context.
func TestBadUsersFieldNoContext(test *testing.T) {
    testCases := []struct{ request any;  }{
        { &struct{ Users CourseUsers }{} },
        { &struct{ User TargetUserSelfOrGrader }{} },
    };

    for i, testCase := range testCases {
        apiErr := checkRequestSpecialFields(nil, testCase.request, "");
        if (apiErr == nil) {
            test.Fatalf("Case %d: Struct with no course context does not return an error: '%+v'.",
                    i, testCase.request);
        }

        if (apiErr.Locator != "-311") {
            test.Fatalf("Case %d: Struct with no course context does not return an error with locator '-311', found '%s': '%+v.",
                    i, apiErr.Locator, testCase.request);
        }
    }
}

// Test CourseUsers and TargetUserSelfOrGrader.
// Users are not exported.
func TestBadUsersFieldNotExported(test *testing.T) {
    testCases := []struct{ request any;  }{
        {
            &struct{ APIRequestCourseUserContext; MinRoleStudent; users CourseUsers }{
                APIRequestCourseUserContext: APIRequestCourseUserContext{
                    CourseID: "course101",
                    UserEmail: "student@test.com",
                    UserPass: studentPass,
                },
            },
        },
        {
            &struct{ APIRequestCourseUserContext; MinRoleStudent; targetUser TargetUserSelfOrGrader }{
                APIRequestCourseUserContext: APIRequestCourseUserContext{
                    CourseID: "course101",
                    UserEmail: "student@test.com",
                    UserPass: studentPass,
                },
            },
        },
    };

    for i, testCase := range testCases {
        apiErr := ValidateAPIRequest(nil, testCase.request, "");
        if (apiErr == nil) {
            test.Fatalf("Case %d: Struct with non-exported field does not return an error: '%+v'.",
                    i, testCase.request);
        }

        if (apiErr.Locator != "-312") {
            test.Fatalf("Case %d: Struct with non-exported field does not return an error with locator '-312', found '%s': '%v.",
                    i, apiErr.Locator, apiErr);
        }
    }
}

func TestBadCourseUsersFieldFailGetUsers(test *testing.T) {
    type goodCourseUsers struct {
        APIRequestCourseUserContext
        MinRoleStudent

        Users CourseUsers
    }

    request := goodCourseUsers{
        APIRequestCourseUserContext: standardCourseContext,
    };

    // First, validate the course context.
    found, apiErr := validateRequestStruct(&request, "");

    if (apiErr != nil) {
        test.Fatalf("Course context validation returned an error when it should be clean: '%v'.", apiErr);
    }

    if (!found) {
        test.Fatalf("Course context validation did not find course context.");
    }

    // Course context is now fine, now make GetUsers fail.
    oldSourcePath := request.Course.SetSourcePathForTesting(filepath.Join(os.DevNull, "course.json"));
    defer func() { request.Course.SetSourcePathForTesting(oldSourcePath); }();

    apiErr = checkRequestSpecialFields(nil, &request, "");
    if (apiErr == nil) {
        test.Fatalf("Error not returned when users fetch failed.");
    }

    expectedLocator := "-313";
    if (apiErr.Locator != expectedLocator) {
        test.Fatalf("Incorrect error locator when user fetch failed. Expcted '%s', found '%s'.",
                expectedLocator, apiErr.Locator);
    }
}

func TestNonEmptyStringField(test *testing.T) {
    testCases := []struct{ request any; errLoc string; jsonName string}{
        {&struct{ APIRequest; Text string }{}, "", ""},

        {&struct{ APIRequest; Text NonEmptyString }{Text: "ZZZ"}, "", "Text"},

        {&struct{ APIRequest; Text NonEmptyString }{}, "-318", "Text"},
        {&struct{ APIRequest; Text NonEmptyString }{Text: ""}, "-318", "Text"},

        {&struct{ APIRequest; Text NonEmptyString `json:"text"`}{}, "-318", "text"},
        {&struct{ APIRequest; Text NonEmptyString `json:"text,omitempty"`}{}, "-318", "text"},
        {&struct{ APIRequest; Text NonEmptyString `json:"foo-bar"`}{}, "-318", "foo-bar"},
        {&struct{ APIRequest; Text NonEmptyString `json:"foo-bar,omitempty"`}{}, "-318", "foo-bar"},
    };

    for i, testCase := range testCases {
        apiErr := ValidateAPIRequest(nil, testCase.request, "");
        if (apiErr != nil) {
            if (testCase.errLoc != "") {
                if (testCase.errLoc != apiErr.Locator) {
                    test.Errorf("Case %d: Incorrect error returned on empty string. Expcted '%s', found '%s'.",
                            i, testCase.errLoc, apiErr.Locator);
                } else {
                    if (testCase.jsonName != apiErr.AdditionalDetails["json-name"]) {
                        test.Errorf("Case %d: Incorrect JSON name returned. Expcted '%s', found '%s'.",
                                i, testCase.jsonName, apiErr.AdditionalDetails["json-name"]);
                    }
                }
            } else {
                test.Errorf("Case %d: Error retutned when it should not be: '%v'.", i, apiErr);
            }
        } else {
            if (testCase.errLoc != "") {
                test.Errorf("Case %d: Error not retutned when it should be.", i);
            }
        }
    }
}

func TestGoodPostFiles(test *testing.T) {
    endpoint := `/test/api/post-files/good`;

    type requestType struct {
        APIRequestCourseUserContext
        MinRoleStudent

        Files POSTFiles
    }

    var tempDir string;

    handler := func(request *requestType) (*string, *APIError) {
        if (len(request.Files.Filenames) != 1) {
            response := fmt.Sprintf("Incorrect number of files. Expected 1, got '%d'.", len(request.Files.Filenames));
            return &response, nil;
        }

        path := filepath.Join(request.Files.TempDir, request.Files.Filenames[0]);
        text, err := util.ReadFile(path);
        if (err != nil) {
            response := fmt.Sprintf("Unable to get files contents from '%s': '%v'.", path, err);
            return &response, nil;
        }

        text = strings.TrimSpace(text);

        expectedText := "a";
        if (text != expectedText) {
            response := fmt.Sprintf("File text not as expected. Expected: '%s', actual: '%s'.", expectedText, text);
            return &response, nil;
        }

        tempDir = request.Files.TempDir;

        return nil, nil;
    }

    routes = append(routes, NewAPIRoute(endpoint, handler));

    paths := []string{
        filepath.Join(config.COURSES_ROOT.Get(), "files", "a.txt"),
    };

    response := SendTestAPIRequestFull(test, endpoint, nil, paths, usr.Admin);
    if (response.Content != nil) {
        test.Fatalf("Handler gave an error: '%s'.", response.Content);
    }

    // Check that the temp dir was cleaned up.
    if (util.PathExists(tempDir)) {
        test.Fatalf("Temp dir was not cleaned up: '%s'.", tempDir);
    }
}

func TestBadPostFilesFieldNotExported(test *testing.T) {
    // Files are not exported.
    type badRequestType struct {
        APIRequestCourseUserContext
        MinRoleStudent

        files POSTFiles
    }

    request := badRequestType{
        APIRequestCourseUserContext: standardCourseContext,
    };

    apiErr := ValidateAPIRequest(nil, &request, "");
    if (apiErr == nil) {
        test.Fatalf("Struct with non-exported files does not return an error,");
    }

    expectedLocator := "-314";
    if (apiErr.Locator != expectedLocator) {
        test.Fatalf("Struct with non-exported files does not return an error with the correct locator. Expcted '%s', found '%s'.",
                expectedLocator, apiErr.Locator);
    }
}

func TestBadPostFilesNoFiles(test *testing.T) {
    endpoint := `/test/api/post-files/bad/no-files`;

    type requestType struct {
        APIRequestCourseUserContext
        MinRoleStudent

        Files POSTFiles
    }

    handler := func(request *requestType) (*any, *APIError) {
        return nil, nil;
    }

    routes = append(routes, NewAPIRoute(endpoint, handler));

    paths := []string{};

    response := SendTestAPIRequestFull(test, endpoint, nil, paths, usr.Admin);
    if (response.Success) {
        test.Fatalf("Request did not generate an error: '%v'.", response);
    }

    expectedLocator := "-316";
    if (response.Locator != expectedLocator) {
        test.Fatalf("Error does not have the correct locator. Expcted '%s', found '%s'.",
                expectedLocator, response.Locator);
    }
}

func TestBadPostFilesStoreFail(test *testing.T) {
    endpoint := `/test/api/post-files/bad/store-fail`;

    type requestType struct {
        APIRequestCourseUserContext
        MinRoleStudent

        Files POSTFiles
    }

    handler := func(request *requestType) (*any, *APIError) {
        return nil, nil;
    }

    routes = append(routes, NewAPIRoute(endpoint, handler));

    paths := []string{
        filepath.Join(config.COURSES_ROOT.Get(), "files", "a.txt"),
    };

    // Ensure that storing the files will fail.
    util.SetTempDirForTesting(os.DevNull);
    defer util.SetTempDirForTesting("");

    response := SendTestAPIRequestFull(test, endpoint, nil, paths, usr.Admin);
    if (response.Success) {
        test.Fatalf("Request did not generate an error: '%v'.", response);
    }

    expectedLocator := "-315";
    if (response.Locator != expectedLocator) {
        test.Fatalf("Error does not have the correct locator. Expcted '%s', found '%s'.",
                expectedLocator, response.Locator);
    }
}

func TestTargetUserJSON(test *testing.T) {
    type testType struct {
        target TargetUser
    }

    testCases := []struct{ in string; expected TargetUser; }{
        {`""`, TargetUser{false, "", nil}},
        {`"a"`, TargetUser{false, "a", nil}},
        {`"student@test.com"`, TargetUser{false, "student@test.com", nil}},
        {`"a\"b\"c"`, TargetUser{false, `a"b"c`, nil}},
    };

    for i, testCase := range testCases {
        var target TargetUser;
        err := json.Unmarshal([]byte(testCase.in), &target);
        if (err != nil) {
            test.Errorf("Case %d: Failed to unmarshal: '%v'.", i, err);
            continue;
        }

        if (testCase.expected != target) {
            test.Errorf("Case %d: Result not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.expected, target);
            continue;
        }

        out, err := util.ToJSON(target);
        if (err != nil) {
            test.Errorf("Case %d: Failed to marshal: '%v'.", i, err);
            continue;
        }

        if (testCase.in != out) {
            test.Errorf("Case %d: Remarshal does not produce the same as input. Expected: '%+v', Actual: '%+v'.", i, testCase.in, out);
            continue;
        }
    }
}

func TestTargetUserSelfOrGraderJSON(test *testing.T) {
    type testType struct {
        target TargetUserSelfOrGrader
    }

    testCases := []struct{ in string; expected TargetUserSelfOrGrader; }{
        {`""`, TargetUserSelfOrGrader{false, "", nil}},
        {`"a"`, TargetUserSelfOrGrader{false, "a", nil}},
        {`"student@test.com"`, TargetUserSelfOrGrader{false, "student@test.com", nil}},
        {`"a\"b\"c"`, TargetUserSelfOrGrader{false, `a"b"c`, nil}},
    };

    for i, testCase := range testCases {
        var target TargetUserSelfOrGrader;
        err := json.Unmarshal([]byte(testCase.in), &target);
        if (err != nil) {
            test.Errorf("Case %d: Failed to unmarshal: '%v'.", i, err);
            continue;
        }

        if (testCase.expected != target) {
            test.Errorf("Case %d: Result not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.expected, target);
            continue;
        }

        out, err := util.ToJSON(target);
        if (err != nil) {
            test.Errorf("Case %d: Failed to marshal: '%v'.", i, err);
            continue;
        }

        if (testCase.in != out) {
            test.Errorf("Case %d: Remarshal does not produce the same as input. Expected: '%+v', Actual: '%+v'.", i, testCase.in, out);
            continue;
        }
    }
}

func TestTargetUserSelfOrGrader(test *testing.T) {
    type requestType struct {
        APIRequestCourseUserContext
        MinRoleOther

        User TargetUserSelfOrGrader
    }

    users, err := grader.GetCourse("course101").GetUsers();
    if (err != nil) {
        test.Fatalf("Failed to get users: '%v'.", err);
    }

    testCases := []struct{ role usr.UserRole; target string; permError bool; expected TargetUserSelfOrGrader; }{
        // Self.
        {usr.Student, "",                 false, TargetUserSelfOrGrader{true, "student@test.com", users["student@test.com"]}},
        {usr.Student, "student@test.com", false, TargetUserSelfOrGrader{true, "student@test.com", users["student@test.com"]}},
        {usr.Grader,  "",                 false, TargetUserSelfOrGrader{true, "grader@test.com", users["grader@test.com"]}},
        {usr.Grader,  "grader@test.com",  false, TargetUserSelfOrGrader{true, "grader@test.com", users["grader@test.com"]}},

        // Other.
        {usr.Other,   "student@test.com", true,  TargetUserSelfOrGrader{true, "student@test.com", users["student@test.com"]}},
        {usr.Student, "grader@test.com",  true,  TargetUserSelfOrGrader{true, "grader@test.com", users["grader@test.com"]}},
        {usr.Grader,  "student@test.com", false, TargetUserSelfOrGrader{true, "student@test.com", users["student@test.com"]}},
        {usr.Admin,   "student@test.com", false, TargetUserSelfOrGrader{true, "student@test.com", users["student@test.com"]}},
        {usr.Owner,   "student@test.com", false, TargetUserSelfOrGrader{true, "student@test.com", users["student@test.com"]}},

        // Not found.
        {usr.Grader, "ZZZ", false, TargetUserSelfOrGrader{false, "ZZZ", nil}},
    };

    for i, testCase := range testCases {
        request := requestType{
            APIRequestCourseUserContext: APIRequestCourseUserContext{
                CourseID: "course101",
                UserEmail: usr.GetRoleString(testCase.role) + "@test.com",
                UserPass: util.Sha256HexFromString(usr.GetRoleString(testCase.role)),
            },
            User: TargetUserSelfOrGrader{
                Email: testCase.target,
            },
        };

        apiErr := ValidateAPIRequest(nil, &request, "");
        if (apiErr != nil) {
            if (testCase.permError) {
                expectedLocator := "-319";
                if (expectedLocator != apiErr.Locator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, apiErr.Locator);
                }
            } else {
                test.Errorf("Case %d: Failed to validate request: '%v'.", i, apiErr);
            }

            continue;
        }

        if (!reflect.DeepEqual(testCase.expected, request.User)) {
            test.Errorf("Case %d: Result not as expected. Expcted '%+v', found '%+v'.",
                    i, testCase.expected, request.User);
        }
    }
}

func TestTargetUser(test *testing.T) {
    type requestType struct {
        APIRequestCourseUserContext
        MinRoleOther

        User TargetUser
    }

    users, err := grader.GetCourse("course101").GetUsers();
    if (err != nil) {
        test.Fatalf("Failed to get users: '%v'.", err);
    }

    testCases := []struct{ role usr.UserRole; target string; expected TargetUser; }{
        {usr.Student, "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
        {usr.Grader,  "grader@test.com",  TargetUser{true, "grader@test.com", users["grader@test.com"]}},

        {usr.Student, "", TargetUser{}},
        {usr.Grader,  "", TargetUser{}},

        {usr.Other,   "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
        {usr.Student, "grader@test.com",  TargetUser{true, "grader@test.com", users["grader@test.com"]}},
        {usr.Grader,  "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
        {usr.Admin,   "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
        {usr.Owner,   "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},

        // Not found.
        {usr.Grader, "ZZZ", TargetUser{false, "ZZZ", nil}},
    };

    for i, testCase := range testCases {
        request := requestType{
            APIRequestCourseUserContext: APIRequestCourseUserContext{
                CourseID: "course101",
                UserEmail: usr.GetRoleString(testCase.role) + "@test.com",
                UserPass: util.Sha256HexFromString(usr.GetRoleString(testCase.role)),
            },
            User: TargetUser{
                Email: testCase.target,
            },
        };

        apiErr := ValidateAPIRequest(nil, &request, "");
        if (apiErr != nil) {
            if (testCase.target == "") {
                expectedLocator := "-320";
                if (expectedLocator != apiErr.Locator) {
                    test.Errorf("Case %d: Incorrect error returned on empty string. Expcted '%s', found '%s'.",
                            i, expectedLocator, apiErr.Locator);
                }
            } else {
                test.Errorf("Case %d: Failed to validate request: '%v'.", i, apiErr);
            }

            continue;
        }

        if (!reflect.DeepEqual(testCase.expected, request.User)) {
            test.Errorf("Case %d: Result not as expected. Expcted '%+v', found '%+v'.",
                    i, testCase.expected, request.User);
        }
    }
}
