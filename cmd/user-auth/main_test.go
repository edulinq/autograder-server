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
	// CMDs always succeed in user authentication, regardless of credentials, so only one test case is needed.
	testCase := struct {
		cmd.CommonCMDTestCase
		targetEmail string
		targetPass  string
	}{
		CommonCMDTestCase: cmd.CommonCMDTestCase{
			ExpectedStdout: expected_auth_output,
		},
		targetEmail: "course-student@test.edulinq.org",
		targetPass:  "course-student",
	}

	args := []string{
		testCase.targetEmail,
		testCase.targetPass,
	}

	cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, "")
}

const expected_auth_output = `{
    "success": true
}
`
