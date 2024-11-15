package cmd

import (
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

// Test if a CMD can start and stop their own server when the primary server isn't being run.
// Since all CMDs use the same infrastructure, we only need to test it for one CMD.
func TestCMDStartsServer(test *testing.T) {
	config.UNIT_TESTING_MODE.Set(true)
	defer util.RemoveRecordedTempDirs()

	expectedSubstrings := []struct {
		expectedSubstring string
	}{
		{`API Server Started.`},
		{`Unix Socket Server Started.`},
		{`API Server Stopped.`},
		{`Unix Socket Server Stopped.`},
	}

	_, stderr, exitCode, err := RunCMDTest(test, userList, []string{}, log.GetTextLevel())
	if err != nil {
		test.Errorf("CMD run returned an error: '%v'.", err)
	}

	if exitCode != 0 {
		test.Errorf("Unexpected exit code. Expected: '0', Actual: '%d'.", exitCode)
	}

	for _, testCase := range expectedSubstrings {
		if !strings.Contains(stderr, testCase.expectedSubstring) {
			test.Errorf("Expected substring '%s' not found in stderr: '%s'.", testCase.expectedSubstring, stderr)
		}
	}
}

func userList() {
	MustHandleCMDRequestAndExit(`users/list`, users.ListRequest{}, users.ListResponse{})
}
