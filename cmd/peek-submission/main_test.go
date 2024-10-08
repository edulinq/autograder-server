package main

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
)

var testCases = []struct {
	cmd.CommonCMDTestCases
	targetEmail      string
	courseID         string
	assignmentID     string
	targetSubmission string
}{
	{cmd.CommonCMDTestCases{0, latestSubmission, ""}, "course-student@test.edulinq.org", "course101", "hw0", ""},
	{cmd.CommonCMDTestCases{0, specificSubmissionShort, ""}, "course-student@test.edulinq.org", "course101", "hw0", "1697406272"},
	{cmd.CommonCMDTestCases{0, specificSubmissionLong, ""}, "course-student@test.edulinq.org", "course101", "hw0", "course101::hw0::student@test.com::1697406256"},

	{cmd.CommonCMDTestCases{0, noSubmission, ""}, "course-admin@test.edulinq.org", "course101", "hw0", ""},
	{cmd.CommonCMDTestCases{0, incorrectSubmission, ""}, "course-student@test.edulinq.org", "course101", "hw0", "ZZZ"},

	{cmd.CommonCMDTestCases{2, "", incorrectCourse}, "course-student@test.edulinq.org", "ZZZ", "hw0", ""},
	{cmd.CommonCMDTestCases{2, "", incorrectAssignment}, "course-student@test.edulinq.org", "course101", "zzz", ""},
}

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestPeekBase(test *testing.T) {
	for i, testCase := range testCases {
		args := []string{
			testCase.targetEmail,
			testCase.courseID,
			testCase.assignmentID,
			testCase.targetSubmission,
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCases, fmt.Sprintf("Case %d: ", i))

		cmd.RunVerboseCMDTests(test, main, args, fmt.Sprintf("Case %d: ", i))
	}
}
