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
			},
			endpoint: "ZZZ",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `Invalid parameter format: missing a colon. Expected format is 'key:value', e.g., 'id:123'. | {"parameter":["course-idcourse101"]}`,
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

		// Base functionality for each supported endpoint.
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "courses/assignments/get",
			parameters: []string{
				"course-id:course101",
				"assignment-id:hw0",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "courses/assignments/list",
			parameters: []string{
				"course-id:course101",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "courses/users/drop",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "courses/users/get",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "courses/users/list",
			parameters: []string{
				"course-id:course101",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "courses/assignments/submissions/fetch/course/attempts",
			parameters: []string{
				"course-id:course101",
				"assignment-id:hw0",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "courses/assignments/submissions/fetch/course/scores",
			parameters: []string{
				"course-id:course101",
				"assignment-id:hw0",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "courses/assignments/submissions/fetch/user/attempt",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
				"assignment-id:hw0",
				"target-submission:1697406256",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "courses/assignments/submissions/fetch/user/history",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
				"assignment-id:hw0",
				"target-submission:1697406256",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
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
				IgnoreStdout: true,
			},
			endpoint: "courses/assignments/submissions/remove",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
				"assignment-id:hw0",
				"target-submission:1697406256",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "lms/user/get",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "logs/query",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "metadata/describe",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "users/get",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
			},
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
			endpoint: "users/list",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				IgnoreStdout: true,
			},
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
