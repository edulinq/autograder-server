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

func TestPeekBase(test *testing.T) {
	testCases := []struct {
		targetEmail        string
		courseID           string
		assignmentID       string
		targetSubmission   string
		expectedSubmission string
		expectedExitCode   int
		expectedStdout     string
		expectedStderr     string
	}{
		{"course-student@test.edulinq.org", "course101", "hw0", "", "1697406272", 0, latestSubmission, ""},
		{"course-student@test.edulinq.org", "course101", "hw0", "1697406272", "1697406272", 0, specificSubmissionShort, ""},
		{"course-student@test.edulinq.org", "course101", "hw0", "course101::hw0::student@test.com::1697406256", "1697406256", 0, specificSubmissionLong, ""},

		{"course-admin@test.edulinq.org", "course101", "hw0", "", "", 0, noSubmission, ""},
		{"course-student@test.edulinq.org", "course101", "hw0", "ZZZ", "", 0, incorrectSubmission, ""},

		{"course-student@test.edulinq.org", "ZZZ", "hw0", "", "", 2, "Could not find course: 'ZZZ'.\n", ""},
		{"course-student@test.edulinq.org", "course101", "zzz", "", "", 2, "Could not find assignment: 'zzz'.\n", ""},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.targetEmail,
			testCase.courseID,
			testCase.assignmentID,
			testCase.targetSubmission,
		}

		commonCases := cmd.CommonCMDTestCases{
			ExpectedExitCode: testCase.expectedExitCode,
			ExpectedStdout:   testCase.expectedStdout,
			ExpectedStderr:   testCase.expectedStderr,
		}

		cmd.RunCommonCMDTests(test, main, args, commonCases, fmt.Sprintf("Case %d: ", i))
	}
}
