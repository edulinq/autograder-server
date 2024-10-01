package main

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestServerUserListBase(test *testing.T) {
	serverUsersMap := db.MustGetServerUsers()

	serverUsers := make([]*model.ServerUser, 0, len(serverUsersMap))
	for _, user := range serverUsersMap {
		serverUsers = append(serverUsers, user)
	}

	expectedServerUserInfos := core.MustNewServerUserInfos(serverUsers)

	testCases := []struct {
		expectedExitCode int
		expectedNumUsers int
		table            bool
	}{
		{0, 10, false},
		{0, 10, true},
	}

	for i, testCase := range testCases {
		args := []string{}

		if testCase.table {
			args = append(args, "--table")
		}

		stdout, err := cmd.RunCommonCMDTests(test, main, args, testCase.expectedExitCode, i)
		if err != nil {
			test.Errorf("Error running common CMD tests: '%v'.", err)
			continue
		}

		if testCase.table {
			oldStdout := os.Stdout
			read, write, err := os.Pipe()
			if err != nil {
				test.Errorf("Case %d: Failed to create pipe: '%v'.", i, err)
			}

			os.Stdout = write

			cmd.ListServerUsersTable(expectedServerUserInfos)

			write.Close()

			os.Stdout = oldStdout

			var outputBuffer bytes.Buffer
			outputBuffer.ReadFrom(read)

			capturedOutput := outputBuffer.String()

			if capturedOutput != stdout {
				test.Errorf("Case %d: Table server user infos do not match. Expected: '%s', Actual: '%s'.", i, capturedOutput, stdout)
			}

		} else {
			var responseContent users.ListResponse
			util.MustJSONFromString(stdout, &responseContent)

			if !reflect.DeepEqual(expectedServerUserInfos, responseContent.Users) {
				test.Errorf("Case %d: Server user infos do not match. Expected: '%v', Actual: '%v'.", i, expectedServerUserInfos, responseContent.Users)
			}
		}
	}
}
