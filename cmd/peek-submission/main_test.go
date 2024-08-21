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
		targetEmail         string
		courseID            string
		assignmentID        string
		targetSubmission    string
		expectedSubmimssion string
	}{
		{"student@test.com", "course101", "hw0", "", "1697406272"},
		{"student@test.com", "course101", "hw0", "1697406272", "1697406272"},
		{"student@test.com", "course101", "hw0", "course101::hw0::student@test.com::1697406256", "1697406256"},

		{"admin@test.com", "course101", "hw0", "", ""},
		{"student@test.com", "course101", "hw0", "ZZZ", ""},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.targetEmail,
			testCase.courseID,
			testCase.assignmentID,
			testCase.targetSubmission,
		}

		stdout, stderr, err := cmd.RunCMDTest(test, main, args)

		if err != nil {
			test.Errorf("Case %d: CMD run returned an error: '%v'.", i, err)
			continue
		}

		if len(stderr) > 0 {
			test.Errorf("Case %d: CMD has content in stderr: '%s'.", i, stderr)
			continue
		}

		var responseContent submissions.FetchUserPeekResponse
		util.MustJSONFromString(stdout, &responseContent)

		expectedHasSubmission := (testCase.expectedSubmimssion != "")
		actualHasSubmission := responseContent.FoundSubmission
		if expectedHasSubmission != actualHasSubmission {
			test.Errorf("Case %d: Incorrect submission existence. Expected: '%v', Actual: '%v'.", i, expectedHasSubmission, actualHasSubmission)
			continue
		}

		if !actualHasSubmission {
			continue
		}

		if testCase.expectedSubmimssion != responseContent.GradingInfo.ShortID {
			test.Errorf("Case %d: Incorrect submission ID. Expected: '%s', Actual: '%s'.", i, testCase.expectedSubmimssion, responseContent.GradingInfo.ShortID)
			continue
		}
	}
}
