package main

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
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
		expectedExitCode    int
		expectedLocator     string
	}{
		{"course-student@test.edulinq.org", "course101", "hw0", "", "1697406272", 0, ""},
		{"course-student@test.edulinq.org", "course101", "hw0", "1697406272", "1697406272", 0, ""},
		{"course-student@test.edulinq.org", "course101", "hw0", "course101::hw0::student@test.com::1697406256", "1697406256", 0, ""},

		{"course-admin@test.edulinq.org", "course101", "hw0", "", "", 0, ""},
		{"course-student@test.edulinq.org", "course101", "hw0", "ZZZ", "", 0, ""},

		{"course-student@test.edulinq.org", "ZZZ", "hw0", "", "", 2, "-018"},
		{"course-student@test.edulinq.org", "course101", "ZZZ", "", "", 2, "-022"},
		{"course-student@test.edulinq.org", "ZZZ", "ZZZ", "", "", 2, "-018"},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.targetEmail,
			testCase.courseID,
			testCase.assignmentID,
			testCase.targetSubmission,
		}

		oldExitCode := util.GetLastExitCode()
		defer func() {
			util.SetExitCode(oldExitCode)
		}()

		util.ShouldExit = false

		stdout, stderr, err := cmd.RunCMDTest(test, main, args)
		if err != nil {
			test.Errorf("Case %d: CMD run returned an error: '%v'.", i, err)
			continue
		}

		if len(stderr) > 0 {
			test.Errorf("Case %d: CMD has content in stderr: '%s'.", i, stderr)
			continue
		}

		actualExitCode := util.GetLastExitCode()
		if actualExitCode != testCase.expectedExitCode {
			test.Errorf("Case %d: Unexpected exit code. Expected: %d, Actual: %d", i, testCase.expectedExitCode, actualExitCode)
			continue
		}

		var response core.APIResponse
		util.MustJSONFromString(stdout, &response)

		if response.Locator != testCase.expectedLocator {
			test.Errorf("Case %d: Unexpected locator code. Expected: %s, Actual: %s", i, testCase.expectedLocator, response.Locator)
			continue
		}

		if testCase.expectedExitCode != 0 {
			continue
		}

		var responseContent submissions.FetchUserPeekResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

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
