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
		cmd.CommonCMDTestCase
		targetEmail      string
		courseID         string
		assignmentID     string
		targetSubmission string
	}{
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			targetEmail:  "course-student@test.edulinq.org",
			courseID:     "course101",
			assignmentID: "hw0",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			targetEmail:      "course-student@test.edulinq.org",
			courseID:         "course101",
			assignmentID:     "hw0",
			targetSubmission: "1697406272",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			targetEmail:      "course-student@test.edulinq.org",
			courseID:         "course101",
			assignmentID:     "hw0",
			targetSubmission: "course101::hw0::student@test.com::1697406272",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: NO_SUBMISSION,
			},
			targetEmail:  "course-admin@test.edulinq.org",
			courseID:     "course101",
			assignmentID: "hw0",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedStdout: INCORRECT_SUBMISSION,
			},
			targetEmail:      "course-student@test.edulinq.org",
			courseID:         "course101",
			assignmentID:     "hw0",
			targetSubmission: "ZZZ",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedExitCode:        1,
				ExpectedStderrSubstring: `Could not find course: 'ZZZ'.`,
				ExpectEmptyStdout:       true,
			},
			targetEmail:  "course-student@test.edulinq.org",
			courseID:     "ZZZ",
			assignmentID: "hw0",
		},
		{
			CommonCMDTestCase: cmd.CommonCMDTestCase{
				ExpectedExitCode:        1,
				ExpectedStderrSubstring: `Could not find assignment: 'zzz'.`,
				ExpectEmptyStdout:       true,
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

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCase, fmt.Sprintf("Case %d: ", i))
	}
}
