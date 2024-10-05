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

func TestServerUserListBase(test *testing.T) {
	testCases := []struct {
		expectedExitCode int
		table            bool
		expectedStdout   string
		expectedStderr   string
	}{
		{0, false, expectedServerUserList, ""},
		{0, true, expectedServerUserListTable, ""},
	}

	for i, testCase := range testCases {
		args := []string{}

		if testCase.table {
			args = append(args, "--table")
		}

		commonCases := cmd.CommonCMDTestCases{
			ExpectedExitCode: testCase.expectedExitCode,
			ExpectedStdout:   testCase.expectedStdout,
			ExpectedStderr:   testCase.expectedStderr,
		}

		cmd.RunCommonCMDTests(test, main, args, commonCases, fmt.Sprintf("Case %d: ", i))
	}
}
