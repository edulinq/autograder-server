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

func TestAssignmentListBase(test *testing.T) {
	testCases := []struct {
		CommonCMDTestCase cmd.CommonCMDTestCase
		courseID          string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_ASSIGNMENT_LIST,
			},
			courseID: "course101",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedExitCode:        1,
				ExpectedStderrSubstring: `Could not find course: 'ZZZ'`,
				ExpectEmptyStdout:       true,
			},
			courseID: "ZZZ",
		},
	}

	for i, testCase := range testCases {
		args := []string{testCase.courseID}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
