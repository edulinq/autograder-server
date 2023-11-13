package core

import (
    "fmt"
    "testing"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var studentPass string = util.Sha256HexFromString("student");

var standardCourseContext APIRequestCourseUserContext = APIRequestCourseUserContext{
    CourseID: "course101",
    UserEmail: "student@test.com",
    UserPass: studentPass,
};

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
    testCases := []struct{value any; role model.UserRole}{
        {struct{}{}, model.RoleUnknown},
        {struct{int}{}, model.RoleUnknown},

        {struct{MinRoleOwner}{}, model.RoleOwner},
        {struct{MinRoleAdmin}{}, model.RoleAdmin},
        {struct{MinRoleGrader}{}, model.RoleGrader},
        {struct{MinRoleStudent}{}, model.RoleStudent},
        {struct{MinRoleOther}{}, model.RoleOther},

        {struct{MinRoleOwner; MinRoleOther}{}, model.RoleOwner},
        {struct{MinRoleAdmin; MinRoleOther}{}, model.RoleAdmin},
        {struct{MinRoleGrader; MinRoleOther}{}, model.RoleGrader},
        {struct{MinRoleStudent; MinRoleOther}{}, model.RoleStudent},

        {struct{MinRoleOther; MinRoleOwner}{}, model.RoleOwner},
        {struct{MinRoleOther; MinRoleAdmin}{}, model.RoleAdmin},
        {struct{MinRoleOther; MinRoleGrader}{}, model.RoleGrader},
        {struct{MinRoleOther; MinRoleStudent}{}, model.RoleStudent},
    };

    for i, testCase := range testCases {
        role, hasRole := getMaxRole(testCase.value);

        if (testCase.role == model.RoleUnknown) {
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

var validBaseAPIRequestTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
    baseAPIRequestTestCase{
        Payload: fmt.Sprintf(`{"course-id": "course101", "assignment-id": "hw0", "user-email": "student@test.com", "user-pass": "%s"}`, studentPass),
        testValues: testValues{A: "", B: 0},
    },
};

var invalidBaseAPIRequestTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
    baseAPIRequestTestCase{Payload: "{}"},
    baseAPIRequestTestCase{Payload: fmt.Sprintf(`{"assignment-id": "hw0", "user-email": "student@test.com", "user-pass": "%s"}`, studentPass)},
    baseAPIRequestTestCase{Payload: fmt.Sprintf(`{"course-id": "course101", "user-email": "student@test.com", "user-pass": "%s"}`, studentPass)},
    baseAPIRequestTestCase{Payload: fmt.Sprintf(`{"course-id": "course101", "assignment-id": "hw0", "user-pass": "%s"}`, studentPass)},
    baseAPIRequestTestCase{Payload: `{"course-id": "course101", "assignment-id": "hw0", "user-email": "student@test.com"}`},
};

var invalidJSONTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
    baseAPIRequestTestCase{Payload: ""},
    baseAPIRequestTestCase{Payload: "{"},
    baseAPIRequestTestCase{Payload: `{course-id": "course101", "assignment-id": "hw0"}`},
    baseAPIRequestTestCase{Payload: `{course-id: "course101", "assignment-id": "hw0"}`},
    baseAPIRequestTestCase{Payload: `{"course-id": course101, "assignment-id": "hw0"}`},
    baseAPIRequestTestCase{Payload: `{"course-id": "course101" "assignment-id": "hw0"}`},
    baseAPIRequestTestCase{Payload: `{"course-id": "course101", "assignment-id": "hw0}`},
};
