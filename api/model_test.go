package api

import (
    "fmt"
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

// Make sure that common setup is done.
func TestMain(suite *testing.M) {
    config.EnableTestingMode(false, true);

    err := grader.LoadCourses();
    if (err != nil) {
        fmt.Printf("Failed to load test courses: '%v'.", err);
        os.Exit(1);
    }

    os.Exit(suite.Run())
}

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

        err = ValidateAPIRequest(request);
        if (err == nil) {
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

        err = ValidateAPIRequest(request);
        if (err != nil) {
            test.Errorf("Case %d: Failed to validate request: '%v'.", i, err);
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

var validBaseAPIRequestTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
    baseAPIRequestTestCase{
        Payload: `{"course-id": "COURSE101", "assignment-id": "hw0", "user-email": "student@test.com", "user-pass": "student"}`,
        testValues: testValues{A: "", B: 0},
    },
};

var invalidBaseAPIRequestTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
    baseAPIRequestTestCase{Payload: "{}"},
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
