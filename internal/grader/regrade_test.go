package grader

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestRegradeBase(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	defer db.ResetForTesting()

	// Note that computation of these paths are deferred until test time.
	studentGradingResults := map[string]*model.GradingResult{
		"1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
	}

	testCases := []struct {
		users             []string
		waitForCompletion bool
		numLeft           int
		results           map[string]*model.SubmissionHistoryItem
	}{
		// User with submission, wait
		{
			[]string{"course-student@test.edulinq.org"},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{
				"course-student@test.edulinq.org": studentGradingResults["1697406272"].Info.ToHistoryItem(),
			},
		},

		// Empty users, wait
		{
			[]string{},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty submissions, wait
		{
			[]string{"course-admin@test.edulinq.org"},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{
				"course-admin@test.edulinq.org": nil,
			},
		},

		// User with submission, no wait
		{
			[]string{"course-student@test.edulinq.org"},
			false,
			1,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty users, no wait
		{
			[]string{},
			false,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty submission, no wait
		{
			[]string{"course-admin@test.edulinq.org"},
			false,
			1,
			map[string]*model.SubmissionHistoryItem{},
		},
	}

	assignment := db.MustGetTestAssignment()

	for i, testCase := range testCases {
		db.ResetForTesting()

		options := RegradeOptions{
			Options:           GetDefaultGradeOptions(),
			Users:             testCase.users,
			Assignment:        assignment,
			WaitForCompletion: testCase.waitForCompletion,
		}

		results, numLeft, err := RegradeSubmissions(options)
		if err != nil {
			test.Errorf("Case %d: Failed to regrade submissions: '%v'.", i, err)
			continue
		}

		if testCase.numLeft != numLeft {
			test.Errorf("Case %d: Unexpected number of regrades remaining. Expected: '%d', actual: '%d'.", i, testCase.numLeft, numLeft)
			continue
		}

		failed := CheckAndClearIDs(test, i, testCase.results, results)
		if failed {
			continue
		}

		if !reflect.DeepEqual(testCase.results, results) {
			test.Errorf("Case %d: Unexpected regrade result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.results), util.MustToJSONIndent(results))
			continue
		}
	}
}
