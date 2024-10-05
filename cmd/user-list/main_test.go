package main

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
)

var testCases = []struct {
	expectedExitCode int
	table            bool
	expectedStdout   string
	expectedStderr   string
}{
	{0, false, expectedServerUserList, ""},
	{0, true, expectedServerUserListTable, ""},
}

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestServerUserListBase(test *testing.T) {
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

func TestServerUserListVerbose(test *testing.T) {
	for i, testCase := range testCases {
		args := []string{}

		if testCase.table {
			args = append(args, "--table")
		}

		_, _, _, err := cmd.RunCMDTest(test, main, args)
		if err != nil {
			test.Errorf("Case %d: CMD run returned an error: '%v'.", i, err)
		}
	}
}
