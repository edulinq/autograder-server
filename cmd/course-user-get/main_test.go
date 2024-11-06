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

func TestCourseUserGetBase(test *testing.T) {
	testCases := []struct {
		cmd.CommonCMDTestCase
		targetEmail string
		courseID    string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: USER_COURSE_FOUND,
			},
			targetEmail: "course-student@test.edulinq.org",
			courseID:    "course101",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: USER_NOT_FOUND,
			},
			targetEmail: "ZZZ",
			courseID:    "course101",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: USER_NOT_FOUND,
			},
			targetEmail: "server-admin@test.edulinq.org",
			courseID:    "course101",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedExitCode:        1,
				ExpectedStderrSubstring: `Could not find course: 'ZZZ'."`,
			},
			targetEmail: "course-student@test.edulinq.org",
			courseID:    "ZZZ",
		},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.targetEmail,
			testCase.courseID,
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
