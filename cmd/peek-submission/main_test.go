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
		cmd.CMDTestParameters
		targetEmail      string
		courseID         string
		assignmentID     string
		targetSubmission string
	}{
		{
			CMDTestParameters: cmd.CMDTestParameters{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			targetEmail:  "course-student@test.edulinq.org",
			courseID:     "course101",
			assignmentID: "hw0",
		},
		{
			CMDTestParameters: cmd.CMDTestParameters{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			targetEmail:      "course-student@test.edulinq.org",
			courseID:         "course101",
			assignmentID:     "hw0",
			targetSubmission: "1697406272",
		},
		{
			CMDTestParameters: cmd.CMDTestParameters{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			targetEmail:      "course-student@test.edulinq.org",
			courseID:         "course101",
			assignmentID:     "hw0",
			targetSubmission: "course101::hw0::student@test.com::1697406272",
		},
		{
			CMDTestParameters: cmd.CMDTestParameters{
				ExpectedStdout: NO_SUBMISSION,
			},
			targetEmail:  "course-admin@test.edulinq.org",
			courseID:     "course101",
			assignmentID: "hw0",
		},
		{
			CMDTestParameters: cmd.CMDTestParameters{
				ExpectedStdout: INCORRECT_SUBMISSION,
			},
			targetEmail:      "course-student@test.edulinq.org",
			courseID:         "course101",
			assignmentID:     "hw0",
			targetSubmission: "ZZZ",
		},
		{
			CMDTestParameters: cmd.CMDTestParameters{
				ExpectedExitCode:        2,
				ExpectedStderrSubstring: `"Could not find course: 'ZZZ'."`,
			},
			targetEmail:  "course-student@test.edulinq.org",
			courseID:     "ZZZ",
			assignmentID: "hw0",
		},
		{
			CMDTestParameters: cmd.CMDTestParameters{
				ExpectedExitCode:        2,
				ExpectedStderrSubstring: `"Could not find assignment: 'zzz'."`,
			},
			targetEmail:  "course-student@test.edulinq.org",
			courseID:     "course101",
			assignmentID: "zzz",
		},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.targetEmail,
			testCase.courseID,
			testCase.assignmentID,
			testCase.targetSubmission,
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CMDTestParameters, fmt.Sprintf("Case %d: ", i))
	}
}
