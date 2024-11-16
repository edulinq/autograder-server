package main

import (
	"fmt"
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
// It must remain or be replaced with an equivalent test if removed.
// CMDs always succeed in user authentication, regardless of credentials, so only one test case is needed.
func TestUserAuthBase(test *testing.T) {
	testCase := []struct {
		cmd.CommonCMDTestCase
		targetEmail string
		targetPass  string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: SUCCESS_AUTH,
			},
			targetEmail: "course-student@test.edulinq.org",
			targetPass:  "course-student",
		},
	}

	for i, testCase := range testCase {
		args := []string{
			testCase.targetEmail,
			testCase.targetPass,
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
