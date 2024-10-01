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
		expectedNumUsers int
		expectedExitCode int
		table            bool
	}{
		{10, 0, false},
		{10, 0, true},
	}

	for i, testCase := range testCases {
		args := []string{}
		if testCase.table {
			args = append(args, "--table")
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

		if testCase.table {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd.ListServerUsersTable(expectedServerUserInfos)

			w.Close()

			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			capturedOutput := buf.String()

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
