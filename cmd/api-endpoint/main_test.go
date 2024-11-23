package main

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

// Test to ensure an API endpoint that requires parameters works correctly.
// Since all CMDs use the same infrastructure, only one endpoint needs to be tested.
func TestApiEndpointBase(test *testing.T) {
	testCases := []struct {
		cmd.CommonCMDTestCase
		endpoint     string
		targetEmail  string
		courseID     string
		assignmentID string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			endpoint:     "courses/assignments/submissions/fetch/user/peek",
			targetEmail:  "target-email:course-student@test.edulinq.org",
			courseID:     "course-id:course101",
			assignmentID: "assignment-id:hw0",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `Failed to find the endpoint.`,
				ExpectedExitCode:        1,
			},
			endpoint:     "ZZZ",
			targetEmail:  "target-email:course-student@test.edulinq.org",
			courseID:     "course-id:course101",
			assignmentID: "assignment-id:hw0",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `No course ID specified.`,
				ExpectedExitCode:        1,
			},
			endpoint:     "courses/assignments/submissions/fetch/user/peek",
			targetEmail:  "target-email:course-student@test.edulinq.org",
			courseID:     "ZZZ:course101",
			assignmentID: "assignment-id:hw0",
		},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.endpoint,
			testCase.targetEmail,
			testCase.courseID,
			testCase.assignmentID,
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
