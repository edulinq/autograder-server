package core

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestAPIErrorApplyContextBase(test *testing.T) {
	type testRequestType struct {
		APIRequestAssignmentContext
	}

	testCases := []struct {
		requestContext any
		expected       *APIError
		leaveTimestamp bool
	}{
		// Errors

		{
			requestContext: nil,
			expected:       &APIError{},
		},
		{
			requestContext: "",
			expected:       &APIError{},
		},
		{
			requestContext: 1,
			expected:       &APIError{},
		},
		{
			requestContext: []string{},
			expected:       &APIError{},
		},

		// String Context

		{
			requestContext: "endpoint",
			expected: &APIError{
				Endpoint: "endpoint",
			},
		},

		// Base Fields

		{
			requestContext: testRequestType{
				APIRequestAssignmentContext: APIRequestAssignmentContext{
					APIRequestCourseUserContext: APIRequestCourseUserContext{
						APIRequestUserContext: APIRequestUserContext{
							APIRequest: APIRequest{
								RequestID: "id",
								Endpoint:  "endpoint",
								Sender:    "sender",
								Timestamp: timestamp.FromMSecs(100),
							},
						},
					},
				},
			},
			expected: &APIError{
				RequestID: "id",
				Endpoint:  "endpoint",
				Sender:    "sender",
				Timestamp: timestamp.FromMSecs(100),
			},
			leaveTimestamp: true,
		},

		// User Fields

		{
			requestContext: testRequestType{
				APIRequestAssignmentContext: APIRequestAssignmentContext{
					APIRequestCourseUserContext: APIRequestCourseUserContext{
						APIRequestUserContext: APIRequestUserContext{
							UserEmail: "U",
						},
					},
				},
			},
			expected: &APIError{
				UserEmail: "U",
			},
		},

		// Course Fields

		{
			requestContext: testRequestType{
				APIRequestAssignmentContext: APIRequestAssignmentContext{
					APIRequestCourseUserContext: APIRequestCourseUserContext{
						CourseID: "C",
					},
				},
			},
			expected: &APIError{
				CourseID: "C",
			},
		},

		// Assignment Fields

		{
			requestContext: testRequestType{
				APIRequestAssignmentContext: APIRequestAssignmentContext{
					AssignmentID: "A",
				},
			},
			expected: &APIError{
				AssignmentID: "A",
			},
		},

		// Mixed Fields

		{
			requestContext: testRequestType{
				APIRequestAssignmentContext: APIRequestAssignmentContext{
					AssignmentID: "A",
					APIRequestCourseUserContext: APIRequestCourseUserContext{
						CourseID: "C",
						APIRequestUserContext: APIRequestUserContext{
							UserEmail: "U",
							APIRequest: APIRequest{
								RequestID: "id",
								Endpoint:  "endpoint",
								Sender:    "sender",
							},
						},
					},
				},
			},
			expected: &APIError{
				RequestID:    "id",
				Endpoint:     "endpoint",
				Sender:       "sender",
				UserEmail:    "U",
				CourseID:     "C",
				AssignmentID: "A",
			},
		},

		// Pointer

		{
			requestContext: &testRequestType{
				APIRequestAssignmentContext: APIRequestAssignmentContext{
					AssignmentID: "A",
				},
			},
			expected: &APIError{
				AssignmentID: "A",
			},
		},
	}

	for i, testCase := range testCases {
		apiError := &APIError{}
		applyContext(apiError, testCase.requestContext)

		// Zero any timestamp.
		if !testCase.leaveTimestamp {
			apiError.Timestamp = timestamp.Zero()
		}

		if !reflect.DeepEqual(testCase.expected, apiError) {
			test.Errorf("Case %d: Unexpected result. Expected '%s', Actual: '%s',",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(apiError))
			continue
		}
	}
}
