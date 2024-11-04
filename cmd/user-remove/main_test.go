package main

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestServerUserRemoveBase(test *testing.T) {
	testCases := []struct {
		cmd.CommonCMDTestCase
		targetEmail string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: USER_FOUND,
			},
			targetEmail: "course-student@test.edulinq.org",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: USER_NOT_FOUND,
			},
			targetEmail: "ZZZ",
		},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.targetEmail,
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
