package main

import (
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/exit"
)

func TestMain(suite *testing.M) {
	// Run inside a func so defers will run before exit.Exit().
	code := func() int {
		db.PrepForTestingMain()
		defer db.CleanupTestingMain()

		return suite.Run()
	}()

	exit.Exit(code)
}

// This test ensures a CMD can start its own server.
// Since this is the only test that ensures a CMD can start its own server,
// it must remain or be replaced with an equivalent test if removed.
func TestUserAuthBase(test *testing.T) {
	commonTestCase := cmd.CommonCMDTestCase{
		ExpectedStdout: expected_auth_output,
	}

	args := []string{
		"course-student@test.edulinq.org",
		"course-student",
	}

	// CMDs always succeed in user authentication, regardless of credentials, so only one test case needs to run.
	cmd.RunCommonCMDTests(test, main, args, commonTestCase, "")
}

const expected_auth_output = `{
    "success": true
}
`
