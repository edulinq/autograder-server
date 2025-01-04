package main

import (
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/db"
)

// This test ensures a CMD can start its own server.
// Since this is the only test that ensures a CMD can start its own server,
// it must remain or be replaced with an equivalent test if removed.
func TestCMDStartsOwnServer(test *testing.T) {
	db.PrepForTestingMain()
	defer db.CleanupTestingMain()

	commonTestCase := cmd.CommonCMDTestCase{
		ExpectedStdout: expected_auth_output,
	}

	args := []string{
		"course-student@test.edulinq.org",
		"course-student",
	}

	// CMDs always succeed in user authentication, regardless of credentials, so only one test case needs to run.
	cmd.RunCommonCMDTests(test, main, args, commonTestCase, "", true)
}

const expected_auth_output = `{
    "success": true
}
`
