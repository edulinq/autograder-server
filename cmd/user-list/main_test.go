package userlist

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
)

var testCases = []struct {
	cmd.CommonCMDTestCases
	table bool
}{
	{cmd.CommonCMDTestCases{0, expectedServerUserList, ""}, false},
	{cmd.CommonCMDTestCases{0, expectedServerUserListTable, ""}, true},
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

		cmd.RunVerboseCMDTests(test, main, args, fmt.Sprintf("Case %d: ", i))
	}
}
