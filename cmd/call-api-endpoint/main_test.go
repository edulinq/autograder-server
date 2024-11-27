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

func TestCallApiEndpointBase(test *testing.T) {
	testCases := []struct {
		cmd.CommonCMDTestCase
		endpoint   string
		parameters []string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: SUBMISSION_1697406256,
			},
			endpoint: "courses/assignments/submissions/fetch/user/peek",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
				"assignment-id:hw0",
				"target-submission:1697406256",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `Failed to find the endpoint. | {"endpoint":"ZZZ"}`,
				ExpectedExitCode:        1,
			},
			endpoint: "ZZZ",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
				"assignment-id:hw0",
				"target-submission:1697406256",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `Invalid parameter format: missing a colon. Expected format is 'key:value', e.g., 'target-email:bob@test.edulinq.org'. | {"parameter":["course-idcourse101"]}`,
				ExpectedExitCode:        1,
			},
			endpoint: "courses/assignments/submissions/fetch/user/peek",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-idcourse101",
				"assignment-id:hw0",
				"target-submission:1697406256",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `No course ID specified.`,
				ExpectedExitCode:        1,
			},
			endpoint: "courses/assignments/submissions/fetch/user/peek",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"ZZZ:course101",
				"assignment-id:hw0",
				"target-submission:1697406256",
			},
		},
	}

	for i, testCase := range testCases {
		args := append([]string{testCase.endpoint}, testCase.parameters...)

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
