package api

import (
    "fmt"
    "path/filepath"
    "strings"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func TestValidBaseCourseUserAPIRequests(test *testing.T) {
    testBaseAPIRequests(test, validBaseAPIRequestTestCases, &baseCourseUserAPIRequest{});
}

func TestValidBaseAssignmentAPIRequests(test *testing.T) {
    testBaseAPIRequests(test, validBaseAPIRequestTestCases, &baseAssignmentAPIRequest{});
}

func TestInvalidBaseAssignmentAPIRequests(test *testing.T) {
    for i, testCase := range invalidBaseAPIRequestTestCases {
        var request baseAssignmentAPIRequest;
        err := util.JSONFromString(testCase.Payload, &request);
        if (err != nil) {
            test.Errorf("Case %d: Failed to unmarshal JSON request ('%s'): '%v'.", i, testCase.Payload, err);
            continue;
        }

        apiErr := ValidateAPIRequest(nil, request, "");
        if (apiErr == nil) {
            test.Errorf("Case %d: Invalid request failed to raise an error.", i);
            continue;
        }
    }
}

func TestInvalidJSON(test *testing.T) {
    for i, testCase := range invalidJSONTestCases {
        var request baseAssignmentAPIRequest;
        err := util.JSONFromString(testCase.Payload, &request);
        if (err == nil) {
            test.Errorf("Case %d: Invalid JSON failed to raise an error.", i);
            continue;
        }
    }
}

func TestGetMaxRole(test *testing.T) {
    testCases := []struct{value any; role usr.UserRole}{
        {struct{}{}, usr.Unknown},
        {struct{int}{}, usr.Unknown},

        {struct{MinRoleOwner}{}, usr.Owner},
        {struct{MinRoleAdmin}{}, usr.Admin},
        {struct{MinRoleGrader}{}, usr.Grader},
        {struct{MinRoleStudent}{}, usr.Student},
        {struct{MinRoleOther}{}, usr.Other},

        {struct{MinRoleOwner; MinRoleOther}{}, usr.Owner},
        {struct{MinRoleAdmin; MinRoleOther}{}, usr.Admin},
        {struct{MinRoleGrader; MinRoleOther}{}, usr.Grader},
        {struct{MinRoleStudent; MinRoleOther}{}, usr.Student},

        {struct{MinRoleOther; MinRoleOwner}{}, usr.Owner},
        {struct{MinRoleOther; MinRoleAdmin}{}, usr.Admin},
        {struct{MinRoleOther; MinRoleGrader}{}, usr.Grader},
        {struct{MinRoleOther; MinRoleStudent}{}, usr.Student},
    };

    for i, testCase := range testCases {
        role, hasRole := getMaxRole(testCase.value);

        if (testCase.role == usr.Unknown) {
            if (hasRole) {
                test.Errorf("Case %d: Found a role ('%s') when none was specified.", i, role);
            }

            continue;
        }

        if (role != testCase.role) {
            test.Errorf("Case %d: Role mismatch. Expected: '%s', Actual: '%s'.", i, testCase.role, role);
        }
    }
}

func TestBadCourseUsersFieldNoContext(test *testing.T) {
    // No embeded course context.
    type badCourseUsersNoCourse struct {
        Users CourseUsers
    }

    apiErr := fillRequestSpecialFields(nil, &badCourseUsersNoCourse{}, "");
    if (apiErr == nil) {
        test.Fatalf("Struct with no course context does not return an error,");
    }

    if (apiErr.RequestID != "-541") {
        test.Fatalf("Struct with no course context does not return an error with request id '-541', found '%s'.", apiErr.RequestID);
    }
}

func TestBadCourseUsersFieldNotExported(test *testing.T) {
    // Users are not exported.
    type badCourseUsersNonExported struct {
        APIRequestCourseUserContext
        MinRoleStudent

        users CourseUsers
    }

    request := badCourseUsersNonExported{
        APIRequestCourseUserContext: APIRequestCourseUserContext{
            CourseID: "COURSE101",
            UserEmail: "student@test.com",
            UserPass: studentPass,
        },
    };

    apiErr := ValidateAPIRequest(nil, &request, "");
    if (apiErr == nil) {
        test.Fatalf("Struct with non-exported course users does not return an error,");
    }

    expectedText := "A CourseUsers field must be exported.";
    if (apiErr.InternalText != expectedText) {
        test.Fatalf("Struct with non-exported course users does not return an error with the correct message. Expcted '%s', found '%s'.",
                expectedText, apiErr.InternalText);
    }
}

func TestBadCourseUsersFieldFailGetUsers(test *testing.T) {
    type goodCourseUsers struct {
        APIRequestCourseUserContext
        MinRoleStudent

        Users CourseUsers
    }

    request := goodCourseUsers{
        APIRequestCourseUserContext: APIRequestCourseUserContext{
            CourseID: "COURSE101",
            UserEmail: "student@test.com",
            UserPass: studentPass,
        },
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
    oldSourcePath := request.course.SourcePath;
    defer func() { request.course.SourcePath = oldSourcePath }();
    request.course.SourcePath = "/dev/null/course.json";

    apiErr = fillRequestSpecialFields(nil, &request, "");
    if (apiErr == nil) {
        test.Fatalf("Error not returned when users fetch failed.");
    }

    expectedText := "Failed to fetch embeded users.";
    if (apiErr.InternalText != expectedText) {
        test.Fatalf("Incorrect error message when user fetch failed. Expcted '%s', found '%s'.",
                expectedText, apiErr.InternalText);
    }
}

// TEST -- need negative tests for post files.

func TestGoodPostFiles(test *testing.T) {
    endpoint := `/test/api/post-files/good`;

    type requestType struct {
        APIRequestCourseUserContext
        MinRoleStudent

        Files POSTFiles
    }

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

        return nil, nil;
    }

    routes = append(routes, newAPIRoute(endpoint, handler));

    paths := []string{
        filepath.Join(config.COURSES_ROOT.GetString(), "files", "a.txt"),
    };

    response := sendTestAPIRequestFull(test, endpoint, nil, paths);

    fmt.Println("###");
    fmt.Println(response);
    fmt.Println("###");

    if (response.Content != nil) {
        test.Fatalf("Handler gave an error: '%s'.", response.Content);
    }
}

func testBaseAPIRequests(test *testing.T, testCases []baseAPIRequestTestCase, request getTestValues) {
    for i, testCase := range testCases {
        err := util.JSONFromString(testCase.Payload, &request);
        if (err != nil) {
            test.Errorf("Case %d: Failed to unmarshal JSON request ('%s'): '%v'.", i, testCase.Payload, err);
            continue;
        }

        if (testCase.testValues != request.GetTestValues()) {
            test.Errorf("Case %d: Request values not as expected. Expected: %v, Actual: %v.", i, testCase.testValues, request.GetTestValues());
            continue;
        }

        apiErr := ValidateAPIRequest(nil, request, "");
        if (apiErr != nil) {
            test.Errorf("Case %d: Failed to validate request: '%v'.", i, apiErr);
            continue;
        }
    }
}

type testValues struct {
    A string `json:"a"`
    B int `json:"b"`
}

type getTestValues interface {
    GetTestValues() testValues
}

type baseAPIRequestTestCase struct {
    Payload string
    testValues
}

type baseCourseUserAPIRequest struct {
    APIRequestCourseUserContext
    MinRoleStudent
    testValues
}

func (this *baseCourseUserAPIRequest) GetTestValues() testValues {
    return this.testValues;
}

type baseAssignmentAPIRequest struct {
    APIRequestAssignmentContext
    MinRoleStudent
    testValues
}

func (this *baseAssignmentAPIRequest) GetTestValues() testValues {
    return this.testValues;
}

var studentPass string = util.Sha256HexFromStrong("student");

var validBaseAPIRequestTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
    baseAPIRequestTestCase{
        Payload: fmt.Sprintf(`{"course-id": "COURSE101", "assignment-id": "hw0", "user-email": "student@test.com", "user-pass": "%s"}`, studentPass),
        testValues: testValues{A: "", B: 0},
    },
};

var invalidBaseAPIRequestTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
    baseAPIRequestTestCase{Payload: "{}"},
    baseAPIRequestTestCase{Payload: fmt.Sprintf(`{"assignment-id": "hw0", "user-email": "student@test.com", "user-pass": "%s"}`, studentPass)},
    baseAPIRequestTestCase{Payload: fmt.Sprintf(`{"course-id": "COURSE101", "user-email": "student@test.com", "user-pass": "%s"}`, studentPass)},
    baseAPIRequestTestCase{Payload: fmt.Sprintf(`{"course-id": "COURSE101", "assignment-id": "hw0", "user-pass": "%s"}`, studentPass)},
    baseAPIRequestTestCase{Payload: `{"course-id": "COURSE101", "assignment-id": "hw0", "user-email": "student@test.com"}`},
};

var invalidJSONTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
    baseAPIRequestTestCase{Payload: ""},
    baseAPIRequestTestCase{Payload: "{"},
    baseAPIRequestTestCase{Payload: `{course-id": "COURSE101", "assignment-id": "hw0"}`},
    baseAPIRequestTestCase{Payload: `{course-id: "COURSE101", "assignment-id": "hw0"}`},
    baseAPIRequestTestCase{Payload: `{"course-id": COURSE101, "assignment-id": "hw0"}`},
    baseAPIRequestTestCase{Payload: `{"course-id": "COURSE101" "assignment-id": "hw0"}`},
    baseAPIRequestTestCase{Payload: `{"course-id": "COURSE101", "assignment-id": "hw0}`},
};
