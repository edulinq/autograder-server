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
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// Test CourseUsers, TargetUserSelfOrGrader, and TargetUserSelfOrAdmin.
// No embeded course context.
func TestBadUsersFieldNoContext(test *testing.T) {
    testCases := []struct{ request any;  }{
        { &struct{ Users CourseUsers }{} },
        { &struct{ User TargetUserSelfOrGrader }{} },
        { &struct{ User TargetUserSelfOrAdmin }{} },
    };

    for i, testCase := range testCases {
        apiErr := checkRequestSpecialFields(nil, testCase.request, "");
        if (apiErr == nil) {
            test.Fatalf("Case %d: Struct with no course context does not return an error: '%+v'.",
                    i, testCase.request);
        }

        if (apiErr.Locator != "-025") {
            test.Fatalf("Case %d: Struct with no course context does not return an error with locator '-025', found '%s': '%+v.",
                    i, apiErr.Locator, testCase.request);
        }
    }
}

// Test CourseUsers, TargetUserSelfOrGrader, and TargetUserSelfOrAdmin.
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
        {
            &struct{ APIRequestCourseUserContext; MinRoleStudent; targetUser TargetUserSelfOrAdmin }{
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

        if (apiErr.Locator != "-026") {
            test.Fatalf("Case %d: Struct with non-exported field does not return an error with locator '-026', found '%s': '%v.",
                    i, apiErr.Locator, apiErr);
        }
    }
}

func TestNonEmptyStringField(test *testing.T) {
    testCases := []struct{ request any; errLoc string; jsonName string}{
        {&struct{ APIRequest; Text string }{}, "", ""},

        {&struct{ APIRequest; Text NonEmptyString }{Text: "ZZZ"}, "", "Text"},

        {&struct{ APIRequest; Text NonEmptyString }{}, "-032", "Text"},
        {&struct{ APIRequest; Text NonEmptyString }{Text: ""}, "-032", "Text"},

        {&struct{ APIRequest; Text NonEmptyString `json:"text"`}{}, "-032", "text"},
        {&struct{ APIRequest; Text NonEmptyString `json:"text,omitempty"`}{}, "-032", "text"},
        {&struct{ APIRequest; Text NonEmptyString `json:"foo-bar"`}{}, "-032", "foo-bar"},
        {&struct{ APIRequest; Text NonEmptyString `json:"foo-bar,omitempty"`}{}, "-032", "foo-bar"},
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
        filepath.Join(config.GetCourseImportDir(), "_tests", "files", "a.txt"),
    };

    response := SendTestAPIRequestFull(test, endpoint, nil, paths, model.RoleAdmin);
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

    expectedLocator := "-028";
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

    response := SendTestAPIRequestFull(test, endpoint, nil, paths, model.RoleAdmin);
    if (response.Success) {
        test.Fatalf("Request did not generate an error: '%v'.", response);
    }

    expectedLocator := "-030";
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
        filepath.Join(config.GetCourseImportDir(), "_tests", "files", "a.txt"),
    };

    // Ensure that storing the files will fail.
    util.SetTempDirForTesting(os.DevNull);
    defer util.SetTempDirForTesting("");

    response := SendTestAPIRequestFull(test, endpoint, nil, paths, model.RoleAdmin);
    if (response.Success) {
        test.Fatalf("Request did not generate an error: '%v'.", response);
    }

    expectedLocator := "-029";
    if (response.Locator != expectedLocator) {
        test.Fatalf("Error does not have the correct locator. Expcted '%s', found '%s'.",
                expectedLocator, response.Locator);
    }
}

func TestBadPostFilesFileSizeExceeded(test *testing.T) {
    resetMaxFileSize := config.WEB_MAX_FILE_SIZE.Get()

    // Set size to 1 byte for testing, then reset when done testing. (a.txt is 2 bytes)
    config.WEB_MAX_FILE_SIZE.Set(1.0);
    defer config.WEB_MAX_FILE_SIZE.Set(resetMaxFileSize)

    endpoint := `/test/api/post-files/bad/size-exceeded`;

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
        filepath.Join(config.GetCourseImportDir(), "_tests", "files", "a.txt"),
    };

    response := SendTestAPIRequestFull(test, endpoint, nil, paths, model.RoleAdmin);
    if (response.Success) {
        test.Fatalf("Request did not generate an error: '%v'.", response);
    }

    expectedLocator := "-036";
    if (response.Locator != expectedLocator) {
        test.Fatalf("Error does not have the correct locator. Expected '%s', found '%s'.",
            expectedLocator, response.Locator);
    }
}

func TestTargetUserJSON(test *testing.T) {
    createTargetType := func(targetUser TargetUser) TargetUser {
        return targetUser;
    }

    testTargetUserJSON(test, createTargetType);
}

func TestTargetUserSelfOrGraderJSON(test *testing.T) {
    createTargetType := func(targetUser TargetUser) TargetUserSelfOrGrader {
        return TargetUserSelfOrGrader{targetUser};
    }

    testTargetUserJSON(test, createTargetType);
}

func TestTargetUserSelfOrAdminJSON(test *testing.T) {
    createTargetType := func(targetUser TargetUser) TargetUserSelfOrAdmin {
        return TargetUserSelfOrAdmin{targetUser};
    }

    testTargetUserJSON(test, createTargetType);
}

func testTargetUserJSON[T comparable](test *testing.T, createTargetType func(TargetUser) T) {
    testCases := []struct{ in string; expected T; }{
        {`""`,                 createTargetType(TargetUser{false, "", nil})},
        {`"a"`,                createTargetType(TargetUser{false, "a", nil})},
        {`"student@test.com"`, createTargetType(TargetUser{false, "student@test.com", nil})},
        {`"a\"b\"c"`,          createTargetType(TargetUser{false, `a"b"c`, nil})},
    };

    for i, testCase := range testCases {
        var target T;
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
    createTargetType := func(targetUser TargetUser) TargetUserSelfOrGrader {
        return TargetUserSelfOrGrader{targetUser};
    }

    createRequest := func(role model.UserRole, target string) *testTargetUserSelfOrGraderRequestType {
        return &testTargetUserSelfOrGraderRequestType{
            APIRequestCourseUserContext: APIRequestCourseUserContext{
                CourseID: "course101",
                UserEmail: model.GetRoleString(role) + "@test.com",
                UserPass: util.Sha256HexFromString(model.GetRoleString(role)),
            },
            User: TargetUserSelfOrGrader{
                TargetUser{
                    Email: target,
                },
            },
        };
    }

    isNonSelfPermError := func(role model.UserRole) bool {
        return role < model.RoleGrader;
    };

    testTargetUser(test, createTargetType, createRequest, isNonSelfPermError);
}

type testTargetUserSelfOrGraderRequestType struct {
    APIRequestCourseUserContext
    MinRoleOther

    User TargetUserSelfOrGrader
}

func (this *testTargetUserSelfOrGraderRequestType) GetUser() any {
    return this.User;
}

func TestTargetUserSelfOrAdmin(test *testing.T) {
    createTargetType := func(targetUser TargetUser) TargetUserSelfOrAdmin {
        return TargetUserSelfOrAdmin{targetUser};
    }

    createRequest := func(role model.UserRole, target string) *testTargetUserSelfOrAdminRequestType {
        return &testTargetUserSelfOrAdminRequestType{
            APIRequestCourseUserContext: APIRequestCourseUserContext{
                CourseID: "course101",
                UserEmail: model.GetRoleString(role) + "@test.com",
                UserPass: util.Sha256HexFromString(model.GetRoleString(role)),
            },
            User: TargetUserSelfOrAdmin{
                TargetUser{
                    Email: target,
                },
            },
        };
    }

    isNonSelfPermError := func(role model.UserRole) bool {
        return role < model.RoleAdmin;
    };

    testTargetUser(test, createTargetType, createRequest, isNonSelfPermError);
}

type testTargetUserSelfOrAdminRequestType struct {
    APIRequestCourseUserContext
    MinRoleOther

    User TargetUserSelfOrAdmin
}

func (this *testTargetUserSelfOrAdminRequestType) GetUser() any {
    return this.User;
}

type userGetter interface {
    GetUser() any
}

func testTargetUser[T comparable, V userGetter](test *testing.T,
        createTargetType func(TargetUser) T,
        createRequest func(model.UserRole, string) V,
        isNonSelfPermError func(model.UserRole) bool) {
    users, err := db.GetUsersFromID("course101");
    if (err != nil) {
        test.Fatalf("Failed to get users: '%v'.", err);
    }

    testCases := []struct{ role model.UserRole; target string; permError bool; expected T; }{
        // Self.
        {model.RoleStudent, "",                 false,
                createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
        {model.RoleStudent, "student@test.com", false,
                createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
        {model.RoleGrader,  "",                 false,
                createTargetType(TargetUser{true, "grader@test.com", users["grader@test.com"]})},
        {model.RoleGrader,  "grader@test.com",  false,
                createTargetType(TargetUser{true, "grader@test.com", users["grader@test.com"]})},

        // Other.
        {model.RoleOther,   "student@test.com", isNonSelfPermError(model.RoleOther),
                createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
        {model.RoleStudent, "grader@test.com",  isNonSelfPermError(model.RoleStudent),
                createTargetType(TargetUser{true, "grader@test.com", users["grader@test.com"]})},
        {model.RoleGrader,  "student@test.com", isNonSelfPermError(model.RoleGrader),
                createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
        {model.RoleAdmin,   "student@test.com", isNonSelfPermError(model.RoleAdmin),
                createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
        {model.RoleOwner,   "student@test.com", isNonSelfPermError(model.RoleOwner),
                createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},

        // Not found.
        {model.RoleGrader, "ZZZ", isNonSelfPermError(model.RoleGrader),
                createTargetType(TargetUser{false, "ZZZ", nil})},
        {model.RoleAdmin, "ZZZ", isNonSelfPermError(model.RoleAdmin),
                createTargetType(TargetUser{false, "ZZZ", nil})},
    };

    for i, testCase := range testCases {
        request := createRequest(testCase.role, testCase.target);

        apiErr := ValidateAPIRequest(nil, request, "");
        if (apiErr != nil) {
            if (testCase.permError) {
                expectedLocator := "-033";
                if (expectedLocator != apiErr.Locator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, apiErr.Locator);
                }
            } else {
                test.Errorf("Case %d: Failed to validate request: '%v'.", i, apiErr);
            }

            continue;
        }

        if (!reflect.DeepEqual(testCase.expected, request.GetUser())) {
            test.Errorf("Case %d: Result not as expected. Expcted '%+v', found '%+v'.",
                    i, testCase.expected, request.GetUser());
        }
    }
}

func TestTargetUser(test *testing.T) {
    type requestType struct {
        APIRequestCourseUserContext
        MinRoleOther

        User TargetUser
    }

    users, err := db.GetUsersFromID("course101");
    if (err != nil) {
        test.Fatalf("Failed to get users: '%v'.", err);
    }

    testCases := []struct{ role model.UserRole; target string; expected TargetUser; }{
        {model.RoleStudent, "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
        {model.RoleGrader,  "grader@test.com",  TargetUser{true, "grader@test.com", users["grader@test.com"]}},

        {model.RoleStudent, "", TargetUser{}},
        {model.RoleGrader,  "", TargetUser{}},

        {model.RoleOther,   "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
        {model.RoleStudent, "grader@test.com",  TargetUser{true, "grader@test.com", users["grader@test.com"]}},
        {model.RoleGrader,  "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
        {model.RoleAdmin,   "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
        {model.RoleOwner,   "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},

        // Not found.
        {model.RoleGrader, "ZZZ", TargetUser{false, "ZZZ", nil}},
    };

    for i, testCase := range testCases {
        request := requestType{
            APIRequestCourseUserContext: APIRequestCourseUserContext{
                CourseID: "course101",
                UserEmail: model.GetRoleString(testCase.role) + "@test.com",
                UserPass: util.Sha256HexFromString(model.GetRoleString(testCase.role)),
            },
            User: TargetUser{
                Email: testCase.target,
            },
        };

        apiErr := ValidateAPIRequest(nil, &request, "");
        if (apiErr != nil) {
            if (testCase.target == "") {
                expectedLocator := "-034";
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
