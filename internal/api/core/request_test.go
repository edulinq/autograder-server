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

func TestAPIRequestAssignmentContextValidateBase(test *testing.T) {
	testCases := []struct {
		request         *APIRequestAssignmentContext
		expectedLocator string
	}{
		// Base
		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-admin@test.edulinq.org",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "course101",
				},
				AssignmentID: "hw0",
			},
			"",
		},

		// Errors

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "course101",
				},
				AssignmentID: "hw0",
			},
			"-016",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-admin@test.edulinq.org",
						UserPass:  "",
					},
					CourseID: "course101",
				},
				AssignmentID: "hw0",
			},
			"-017",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-admin@test.edulinq.org",
						UserPass:  "AAA",
					},
					CourseID: "course101",
				},
				AssignmentID: "hw0",
			},
			"-014",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "ZZZ",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "course101",
				},
				AssignmentID: "hw0",
			},
			"-013",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "root",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "course101",
				},
				AssignmentID: "hw0",
			},
			"-051",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-admin@test.edulinq.org",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "",
				},
				AssignmentID: "hw0",
			},
			"-015",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-admin@test.edulinq.org",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "course!!!id",
				},
				AssignmentID: "hw0",
			},
			"-052",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-admin@test.edulinq.org",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "ZZZ",
				},
				AssignmentID: "hw0",
			},
			"-018",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-user@test.edulinq.org",
						UserPass:  util.Sha256HexFromString("server-user"),
					},
					CourseID: "course101",
				},
				AssignmentID: "hw0",
			},
			"-040",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-admin@test.edulinq.org",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "course101",
				},
				AssignmentID: "",
			},
			"-021",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-admin@test.edulinq.org",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "course101",
				},
				AssignmentID: "hw!!!0",
			},
			"-035",
		},

		{
			&APIRequestAssignmentContext{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "server-admin@test.edulinq.org",
						UserPass:  util.Sha256HexFromString("server-admin"),
					},
					CourseID: "course101",
				},
				AssignmentID: "zzz",
			},
			"-022",
		},
	}

	for i, testCase := range testCases {
		err := testCase.request.Validate(nil, baseAssignmentAPIRequest{}, "")
		if err != nil {
			if testCase.expectedLocator == "" {
				test.Errorf("Case %d: Unexpected error ('%s'): '%v'.", i, err.Locator, err)
				continue
			}

			if testCase.expectedLocator != err.Locator {
				test.Errorf("Case %d: Did not get the expected locator. Expected: '%s', Actual: '%s'.", i, testCase.expectedLocator, err.Locator)
				continue
			}

			continue
		}

		if testCase.expectedLocator != "" {
			test.Errorf("Case %d: Did not get expected error: '%s'.", i, testCase.expectedLocator)
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
