package main

import (
	"strconv"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestServerUserListBase(test *testing.T) {
	testCases := []struct {
		expectedNumUsers string
		expectedExitCode int
		filteredUsers    []string
	}{
		{"10", 0, []string{}},
		{"1", 0, []string{"root"}},
		{"2", 0, []string{"root", "course-admin@test.edulinq.org"}},
		{"1", 0, []string{"root", "ZZZ"}},
	}

	for i, testCase := range testCases {
		args := []string{}

		if len(testCase.filteredUsers) > 0 {
			args = append(args, testCase.filteredUsers...)
		}

		stdout, stderr, exitCode, err := cmd.RunCMDTest(test, main, args)
		if err != nil {
			test.Errorf("Case %d: CMD run returned an error: '%v'.", i, err)
			continue
		}

		if len(stderr) > 0 {
			test.Errorf("Case %d: CMD has content in stderr: '%s'.", i, stderr)
			continue
		}

		if testCase.expectedExitCode != exitCode {
			test.Errorf("Case %d: Unexpected exit code. Expected: '%d', Actual: '%d'.", i, testCase.expectedExitCode, exitCode)
			continue
		}

		var response core.APIResponse
		util.MustJSONFromString(stdout, &response)
		var responseContent users.ListResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		expectedUsers, err := strconv.Atoi(testCase.expectedNumUsers)
		if err != nil {
			test.Errorf("Case %d: Failed to convert string to int.", i)
			continue
		}

		actualUsers := len(responseContent.Users)
		if expectedUsers != actualUsers {
			test.Errorf("Case %d: Unexpected number of server users. Expected: '%d', Actual: '%d'.", i, expectedUsers, actualUsers)
			continue
		}
	}
}
