package core

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

var studentPass string = util.Sha256HexFromString("course-student")

var standardUserContext APIRequestUserContext = APIRequestUserContext{
	UserEmail: "course-student@test.edulinq.org",
	UserPass:  studentPass,
}

var standardCourseContext APIRequestCourseUserContext = APIRequestCourseUserContext{
	APIRequestUserContext: standardUserContext,
	CourseID:              "course101",
}

func TestValidBaseCourseUserAPIRequests(test *testing.T) {
	testBaseAPIRequests(test, validBaseAPIRequestTestCases, &baseCourseUserAPIRequest{})
}

func TestValidBaseAssignmentAPIRequests(test *testing.T) {
	testBaseAPIRequests(test, validBaseAPIRequestTestCases, &baseAssignmentAPIRequest{})
}

func TestInvalidBaseAssignmentAPIRequests(test *testing.T) {
	for i, testCase := range invalidBaseAPIRequestTestCases {
		var request baseAssignmentAPIRequest
		err := util.JSONFromString(testCase.Payload, &request)
		if err != nil {
			test.Errorf("Case %d: Failed to unmarshal JSON request ('%s'): '%v'.", i, testCase.Payload, err)
			continue
		}

		apiErr := ValidateAPIRequest(nil, request, "")
		if apiErr == nil {
			test.Errorf("Case %d: Invalid request failed to raise an error.", i)
			continue
		}
	}
}

func TestInvalidJSON(test *testing.T) {
	for i, testCase := range invalidJSONTestCases {
		var request baseAssignmentAPIRequest
		err := util.JSONFromString(testCase.Payload, &request)
		if err == nil {
			test.Errorf("Case %d: Invalid JSON failed to raise an error.", i)
			continue
		}
	}
}

func testBaseAPIRequests(test *testing.T, testCases []baseAPIRequestTestCase, request getTestValues) {
	for i, testCase := range testCases {
		err := util.JSONFromString(testCase.Payload, &request)
		if err != nil {
			test.Errorf("Case %d: Failed to unmarshal JSON request ('%s'): '%v'.", i, testCase.Payload, err)
			continue
		}

		if testCase.testValues != request.GetTestValues() {
			test.Errorf("Case %d: Request values not as expected. Expected: %v, Actual: %v.", i, testCase.testValues, request.GetTestValues())
			continue
		}

		apiErr := ValidateAPIRequest(nil, request, "")
		if apiErr != nil {
			test.Errorf("Case %d: Failed to validate request: '%v'.", i, apiErr)
			continue
		}
	}
}

type testValues struct {
	A string `json:"a"`
	B int    `json:"b"`
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
	MinCourseRoleStudent
	testValues
}

func (this *baseCourseUserAPIRequest) GetTestValues() testValues {
	return this.testValues
}

type baseAssignmentAPIRequest struct {
	APIRequestAssignmentContext
	MinCourseRoleStudent
	testValues
}

func (this *baseAssignmentAPIRequest) GetTestValues() testValues {
	return this.testValues
}

var validBaseAPIRequestTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
	baseAPIRequestTestCase{
		Payload:    fmt.Sprintf(`{"course-id": "course101", "assignment-id": "hw0", "user-email": "course-student@test.edulinq.org", "user-pass": "%s"}`, studentPass),
		testValues: testValues{A: "", B: 0},
	},
}

var invalidBaseAPIRequestTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
	baseAPIRequestTestCase{Payload: "{}"},
	baseAPIRequestTestCase{Payload: fmt.Sprintf(`{"assignment-id": "hw0", "user-email": "course-student@test.edulinq.org", "user-pass": "%s"}`, studentPass)},
	baseAPIRequestTestCase{Payload: fmt.Sprintf(`{"course-id": "course101", "user-email": "course-student@test.edulinq.org", "user-pass": "%s"}`, studentPass)},
	baseAPIRequestTestCase{Payload: fmt.Sprintf(`{"course-id": "course101", "assignment-id": "hw0", "user-pass": "%s"}`, studentPass)},
	baseAPIRequestTestCase{Payload: `{"course-id": "course101", "assignment-id": "hw0", "user-email": "course-student@test.edulinq.org"}`},
}

var invalidJSONTestCases []baseAPIRequestTestCase = []baseAPIRequestTestCase{
	baseAPIRequestTestCase{Payload: ""},
	baseAPIRequestTestCase{Payload: "{"},
	baseAPIRequestTestCase{Payload: `{course-id": "course101", "assignment-id": "hw0"}`},
	baseAPIRequestTestCase{Payload: `{course-id: "course101", "assignment-id": "hw0"}`},
	baseAPIRequestTestCase{Payload: `{"course-id": course101, "assignment-id": "hw0"}`},
	baseAPIRequestTestCase{Payload: `{"course-id": "course101" "assignment-id": "hw0"}`},
	baseAPIRequestTestCase{Payload: `{"course-id": "course101", "assignment-id": "hw0}`},
}
