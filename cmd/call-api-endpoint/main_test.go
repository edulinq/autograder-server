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
		// Errors.
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `Failed to find the endpoint. | {"endpoint":"ZZZ"}`,
				ExpectedExitCode:        1,
				ExpectEmptyStdout:       true,
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
				ExpectEmptyStdout:       true,
			},
			endpoint: "courses/assignments/submissions/fetch/user/peek",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-idcourse101",
				"assignment-id:hw0",
				"target-submission:1697406256",
			},
		},

		// Base functionality for each supported endpoint.
		{
			endpoint: "courses/assignments/get",
			parameters: []string{
				"course-id:course101",
				"assignment-id:hw0",
			},
		},
		{
			endpoint: "courses/assignments/list",
			parameters: []string{
				"course-id:course101",
			},
		},
		{
			endpoint: "courses/users/drop",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
			},
		},
		{
			endpoint: "courses/users/get",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
			},
		},
		{
			endpoint: "courses/users/list",
			parameters: []string{
				"course-id:course101",
			},
		},
		{
			endpoint: "courses/assignments/submissions/fetch/user/peek",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
				"assignment-id:hw0",
				"target-submission:1697406256",
			},
		},
		{
			endpoint: "users/get",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
			},
		},
		{
			endpoint: "users/list",
		},
		{
			endpoint: "users/remove",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
			},
		},
	}

	for i, testCase := range testCases {
		args := append([]string{testCase.endpoint}, testCase.parameters...)

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
