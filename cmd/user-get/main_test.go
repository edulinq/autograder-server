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

func TestServerUserGetBase(test *testing.T) {
	testCases := []struct {
		CommonCMDTestCase cmd.CommonCMDTestCase
		targetEmail       string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_SERVER_USER_GET,
			},
			targetEmail: "server-user@test.edulinq.org",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_UNKNOWN_SERVER_USER_GET,
			},
			targetEmail: "ZZZ",
		},
	}

	for i, testCase := range testCases {
		args := []string{testCase.targetEmail}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
