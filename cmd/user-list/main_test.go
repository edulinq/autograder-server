package main

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
)

var testCases = []struct {
	CommonCMDTestCases cmd.CommonCMDTestCases
	table              bool
}{
	{
		CommonCMDTestCases: cmd.CommonCMDTestCases{
			ExpectedStdout: EXPECTED_SERVER_USER_LIST,
		},
	},
	{
		CommonCMDTestCases: cmd.CommonCMDTestCases{
			ExpectedStdout: EXPECTED_SERVER_USER_LIST_TABLE,
		},
		table: true,
	},
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

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCases, fmt.Sprintf("Case %d: ", i))
	}
}

// Test to ensure that the verbose flag doesn't cause an error.
// Since the verbose flag is common code, we only need to test it for one CMD.
func TestServerUserListVerbose(test *testing.T) {
	args := []string{"--verbose"}

	_, _, _, err := cmd.RunCMDTest(test, main, args)
	if err != nil {
		test.Errorf("CMD run returned an error when testing verbose: '%v'.", err)
	}
}
