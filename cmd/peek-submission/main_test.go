package main

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestPeekBase(test *testing.T) {
	testCases := []struct {
		targetEmail          string
		courseID             string
		assignmentID         string
		targetSubmission     string
		expectedSubmission   string
		expectedExitCode     int
		expectedErrorMessage string
	}{
		{"course-student@test.edulinq.org", "course101", "hw0", "", "1697406272", 0, ""},
		{"course-student@test.edulinq.org", "course101", "hw0", "1697406272", "1697406272", 0, ""},
		{"course-student@test.edulinq.org", "course101", "hw0", "course101::hw0::student@test.com::1697406256", "1697406256", 0, ""},

		{"course-admin@test.edulinq.org", "course101", "hw0", "", "", 0, ""},
		{"course-student@test.edulinq.org", "course101", "hw0", "ZZZ", "", 0, ""},

		{"course-student@test.edulinq.org", "ZZZ", "hw0", "", "", 2, "Could not find course: 'ZZZ'.\n"},
		{"course-student@test.edulinq.org", "course101", "zzz", "", "", 2, "Could not find assignment: 'zzz'.\n"},
		{"course-student@test.edulinq.org", "ZZZ", "zzz", "", "", 2, "Could not find course: 'ZZZ'.\n"},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.targetEmail,
			testCase.courseID,
			testCase.assignmentID,
			testCase.targetSubmission,
		}

		stdout, err := cmd.RunCommonCMDTests(test, main, args, testCase.expectedExitCode, i)
		if err != nil {
			test.Errorf("Error running common CMD tests: '%v'.", err)
			continue
		}

		if testCase.expectedErrorMessage != "" && stdout != testCase.expectedErrorMessage {
			test.Errorf("Case: %d: Unexpected error message. Expected: '%s', Actual: '%s'.", i, testCase.expectedErrorMessage, stdout)
		}

		if testCase.expectedExitCode != 0 {
			continue
		}

		var responseContent submissions.FetchUserPeekResponse
		util.MustJSONFromString(stdout, &responseContent)

		expectedHasSubmission := (testCase.expectedSubmission != "")
		actualHasSubmission := responseContent.FoundSubmission
		if expectedHasSubmission != actualHasSubmission {
			test.Errorf("Case %d: Incorrect submission existence. Expected: '%v', Actual: '%v'.", i, expectedHasSubmission, actualHasSubmission)
			continue
		}

		if !actualHasSubmission {
			continue
		}

		if testCase.expectedSubmission != responseContent.GradingInfo.ShortID {
			test.Errorf("Case %d: Incorrect submission ID. Expected: '%s', Actual: '%s'.", i, testCase.expectedSubmission, responseContent.GradingInfo.ShortID)
			continue
		}
	}
}
