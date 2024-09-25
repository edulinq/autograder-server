package main

import (
	"sort"
	"testing"

	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestServerUserListBase(test *testing.T) {
	testCases := []struct {
		expectedNumUsers int
		expectedExitCode int
	}{
		{10, 0},
	}

	serverUsersMap := db.MustGetServerUsers()

	expectedEmailList := make([]string, 0, len(serverUsersMap))
	for _, user := range serverUsersMap {
		expectedEmailList = append(expectedEmailList, user.Email)
	}

	for i, testCase := range testCases {
		args := []string{}

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

		var responseContent users.ListResponse
		util.MustJSONFromString(stdout, &responseContent)

		actualNumUsers := len(responseContent.Users)
		if testCase.expectedNumUsers != actualNumUsers {
			test.Errorf("Case %d: Unexpected number of server users. Expected: '%d', Actual: '%d'.", i, testCase.expectedNumUsers, actualNumUsers)
			continue
		}

		actualEmailList := make([]string, 0, len(responseContent.Users))
		for _, user := range responseContent.Users {
			actualEmailList = append(actualEmailList, user.Email)
		}

		sort.Strings(expectedEmailList)
		sort.Strings(actualEmailList)

		for email := range expectedEmailList {
			if expectedEmailList[email] != actualEmailList[email] {
				test.Errorf("Case %d: Unexpected server user. Expected: '%s', Actual: '%s'.", i, expectedEmailList[email], actualEmailList[email])
			}
		}
	}
}
