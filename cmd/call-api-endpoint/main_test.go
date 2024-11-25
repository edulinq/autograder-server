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
func TestApiEndpointBase(test *testing.T) {
	testCases := []struct {
		cmd.CommonCMDTestCase
		endpoint     string
		targetEmail  string
		courseID     string
		assignmentID string
		submissionID string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: SUBMISSION_1697406256,
			},
			endpoint:     "courses/assignments/submissions/fetch/user/peek",
			targetEmail:  "target-email:course-student@test.edulinq.org",
			courseID:     "course-id:course101",
			assignmentID: "assignment-id:hw0",
			submissionID: "target-submission:1697406256",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout:   `Failed to find the endpoint 'ZZZ'.`,
				ExpectedExitCode: 1,
			},
			endpoint:     "ZZZ",
			targetEmail:  "target-email:course-student@test.edulinq.org",
			courseID:     "course-id:course101",
			assignmentID: "assignment-id:hw0",
			submissionID: "target-submission:1697406256",
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
			submissionID: "target-submission:1697406256",
		},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.endpoint,
			testCase.targetEmail,
			testCase.courseID,
			testCase.assignmentID,
			testCase.submissionID,
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
