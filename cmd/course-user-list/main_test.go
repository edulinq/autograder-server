package main

import (
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestCourseUserListBase(test *testing.T) {
	commonTestCase := cmd.CommonCMDTestCase{
		ExpectedStdout: EXPECTED_COURSE_USER_LIST_TABLE,
	}

	args := []string{"course101", "--table"}

	cmd.RunCommonCMDTests(test, main, args, commonTestCase, "")
}
