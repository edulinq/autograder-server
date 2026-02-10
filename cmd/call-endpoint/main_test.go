package main

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestCallEndpoint(test *testing.T) {
	apiDescription, err := core.DescribeRoutes(*api.GetRoutes())
	if err != nil {
		test.Fatalf("Failed to get API description: '%v'.", err)
	}

	listRows := make([]string, 0, len(apiDescription.Endpoints))
	for endpoint, _ := range apiDescription.Endpoints {
		listRows = append(listRows, endpoint)
	}
	slices.Sort(listRows)

	testCases := []struct {
		cmd.CommonCMDTestCase
		endpoint   string
		parameters []string
	}{
		// Empty Substring
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `Please enter an endpoint. Use --list to view all endpoints.`,
				ExpectedExitCode:        1,
			},
			endpoint: "",
		},

		// No Matches
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `Failed to find matching endpoint. Use --list to view all endpoints. | {"endpoint-substring":"ZZZ"}`,
				ExpectedExitCode:        1,
			},
			endpoint: "ZZZ",
		},

		// Shorthand
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: util.MustToJSONIndent(apiDescription.Endpoints["metadata/heartbeat"]) + "\n",
			},
			endpoint: "heartbeat",
			parameters: []string{
				"--describe",
			},
		},

		// Multiple Matches
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `Found multiple matching endpoints. Use --list to view all endpoints. | {"endpoint-substring":"sers/list","matching-endpoints":["courses/users/list","users/list"]}`,
				ExpectedExitCode:        1,
			},
			endpoint: "sers/list",
			parameters: []string{
				"--describe",
			},
		},

		// Multiple Matches - Full Match
		// Also matches: courses/users/list
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: util.MustToJSONIndent(apiDescription.Endpoints["users/list"]) + "\n",
			},
			endpoint: "users/list",
			parameters: []string{
				"--describe",
			},
		},

		// Exact Match
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: util.MustToJSONIndent(apiDescription.Endpoints["metadata/heartbeat"]) + "\n",
			},
			endpoint: "metadata/heartbeat",
			parameters: []string{
				"--describe",
			},
		},

		// Exact Match - Multiple
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStderrSubstring: `Failed to find an exact endpoint match. Use --list to view all endpoints. | {"endpoint-substring":"heartbeat"}`,
				ExpectedExitCode:        1,
			},
			endpoint: "heartbeat",
			parameters: []string{
				"--exact-match",
				"--describe",
			},
		},

		// Bad Parameter
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

		// List
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: strings.Join(listRows, "\n") + "\n",
			},
			endpoint: "",
			parameters: []string{
				"--list",
			},
		},

		// Describe
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: util.MustToJSONIndent(apiDescription.Endpoints["users/auth"]) + "\n",
			},
			endpoint: "users/auth",
			parameters: []string{
				"--describe",
			},
		},

		// Input: Bool
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_TESTING_STATS,
			},
			endpoint: "stats/query",
			parameters: []string{
				"type:cpu-usage",
				"use-testing-data:true",
			},
		},

		// Output: Single key with a list.
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_COURSES_ASSIGNMENTS_LIST_TABLE,
			},
			endpoint: "courses/assignments/list",
			parameters: []string{
				"course-id:course101",
				"--table",
			},
		},

		// Output: Single key with a map.
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_COURSES_ASSIGNMENTS_GET_TABLE,
			},
			endpoint: "courses/assignments/get",
			parameters: []string{
				"course-id:course101",
				"assignment-id:hw0",
				"--table",
			},
		},

		// Output: Multiple keys with a list.
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_FETCH_USER_HISTORY_TABLE,
			},
			endpoint: "courses/assignments/submissions/fetch/user/history",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
				"assignment-id:hw0",
				"target-submission:1697406256",
				"--table",
			},
		},

		// Output: Multiple keys with a map.
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_COURSES_USERS_GET_TABLE,
			},
			endpoint: "courses/users/get",
			parameters: []string{
				"target-email:course-student@test.edulinq.org",
				"course-id:course101",
				"--table",
			},
		},
	}

	for i, testCase := range testCases {
		args := append([]string{testCase.endpoint}, testCase.parameters...)

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}

// Test to ensure no panic occurs when converting an API response to a table.
func TestTableConversion(test *testing.T) {
	response := core.APIResponse{
		Content: map[string]any{
			"key": []any{},
		},
	}

	_, _ = cmd.ConvertAPIResponseToTable(response)
}

func TestUpdateEndpointParams(test *testing.T) {
	apiDescription, err := core.DescribeRoutes(*api.GetRoutes())
	if err != nil {
		test.Fatalf("Failed to get API description: '%v'.", err)
	}

	testCases := []struct {
		endpoint       string
		params         map[string]any
		expected       map[string]any
		errorSubstring string
	}{
		// Boolean - True
		{
			endpoint: "stats/query",
			params: map[string]any{
				"use-testing-data": "true",
			},
			expected: map[string]any{
				"use-testing-data": true,
			},
		},

		// Boolean - False
		{
			endpoint: "stats/query",
			params: map[string]any{
				"use-testing-data": "false",
			},
			expected: map[string]any{
				"use-testing-data": false,
			},
		},

		// Boolean - Error
		{
			endpoint: "stats/query",
			params: map[string]any{
				"use-testing-data": "TRUE",
			},
			errorSubstring: "Param 'use-testing-data': Failed to convert boolean",
		},
	}

	for i, testCase := range testCases {
		err := updateParams(apiDescription.Endpoints[testCase.endpoint], testCase.params)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to update params: '%v'.", i, err)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error.", i)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, testCase.params) {
			test.Errorf("Case %d: Params not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(testCase.params))
			continue
		}
	}
}
