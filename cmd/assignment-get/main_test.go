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

func TestAssignmentGetBase(test *testing.T) {
	testCases := []struct {
		cmd.CommonCMDTestCase
		courseID     string
		assignmentID string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: COURSE_ASSIGNMENT_FOUND,
			},
			courseID:     "course101",
			assignmentID: "hw0",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedExitCode:        1,
				ExpectedStderrSubstring: `Could not find course: 'ZZZ'.`,
				ExpectEmptyStdout:       true,
			},
			courseID:     "ZZZ",
			assignmentID: "hw0",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedExitCode:        1,
				ExpectedStderrSubstring: `Could not find assignment: 'zzz'.`,
				ExpectEmptyStdout:       true,
			},
			courseID:     "course101",
			assignmentID: "zzz",
		},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.courseID,
			testCase.assignmentID,
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
