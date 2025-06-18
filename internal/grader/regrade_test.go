package grader

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/jobmanager"
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
		users             []model.CourseUserReference
		waitForCompletion bool
		numLeft           int
		results           map[string]*model.SubmissionHistoryItem
	}{
		// User With Submission, Wait
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{
				"course-student@test.edulinq.org": studentGradingResults["1697406272"].Info.ToHistoryItem(),
			},
		},

		// Empty Users, Wait
		{
			[]model.CourseUserReference{},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty Submissions, Wait
		{
			[]model.CourseUserReference{"course-admin@test.edulinq.org"},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{
				"course-admin@test.edulinq.org": nil,
			},
		},

		// User With Submission, No Wait
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			false,
			1,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty Users, No Wait
		{
			nil,
			false,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},
		{
			[]model.CourseUserReference{},
			false,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},
		{
			[]model.CourseUserReference{"-*"},
			false,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty Submission, No Wait
		{
			[]model.CourseUserReference{"course-admin@test.edulinq.org"},
			false,
			1,
			map[string]*model.SubmissionHistoryItem{},
		},
	}

	assignment := db.MustGetTestAssignment()

	for i, testCase := range testCases {
		db.ResetForTesting()

		options := RegradeOptions{
			GradeOptions: GetDefaultGradeOptions(),
			JobOptions: jobmanager.JobOptions{
				WaitForCompletion: testCase.waitForCompletion,
			},
			RawReferences: testCase.users,
			// TODO: Make this a test case field.
			RegradeAfter:          nil,
			RetainOriginalContext: false,
		}

		results, regradeAfter, numLeft, workErrors, err := Regrade(assignment, options)
		if err != nil {
			test.Errorf("Case %d: Failed to regrade submissions: '%v'.", i, err)
			continue
		}

		if len(workErrors) != 0 {
			test.Errorf("Case %d: Unexpected work errors during regrade: '%s'.", i, util.MustToJSONIndent(workErrors))
			continue
		}

		// TODO: Add a check for regradeAfter.
		if regradeAfter == 0 {
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
