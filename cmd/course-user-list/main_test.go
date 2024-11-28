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

func TestCourseUserListBase(test *testing.T) {
	testCases := []struct {
		CommonCMDTestCase cmd.CommonCMDTestCase
		courseID          string
		table             bool
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_COURSE_USER_LIST,
			},
			courseID: "course101",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: EXPECTED_COURSE_USER_LIST_TABLE,
			},
			table:    true,
			courseID: "course101",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedExitCode:        1,
				ExpectedStderrSubstring: `Could not find course: 'ZZZ'.`,
				ExpectEmptyStdout:       true,
			},
			courseID: "ZZZ",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedExitCode:        1,
				ExpectedStderrSubstring: `Could not find course: 'ZZZ'.`,
				ExpectEmptyStdout:       true,
			},
			table:    true,
			courseID: "ZZZ",
		},
	}

	for i, testCase := range testCases {
		args := []string{testCase.courseID}

		if testCase.table {
			args = append(args, "--table")
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
