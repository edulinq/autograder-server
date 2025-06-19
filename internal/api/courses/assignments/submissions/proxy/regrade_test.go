package proxy

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestRegradeBase(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	db.ResetForTesting()
	defer db.ResetForTesting()

	// Note that computation of these paths are deferred until test time.
	studentGradingResults := map[string]*model.GradingResult{
		"1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
	}

	testCases := []struct {
		options         grader.RegradeOptions
		proxyUser       string
		expectedLocator string
		expected        RegradeResponse
	}{
		// Student, Wait For Completion
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
				RawReferences: []model.CourseUserReference{"student"},
			},
			"course-grader",
			"",
			RegradeResponse{
				RegradeResult: grader.RegradeResult{
					Results: map[string]*model.SubmissionHistoryItem{
						"course-student@test.edulinq.org": studentGradingResults["1697406272"].Info.ToHistoryItem(),
					},
					WorkErrors: map[string]string{},
				},
				Complete:      true,
				ResolvedUsers: []string{"course-student@test.edulinq.org"},
			},
		},

		// Admin, Wait For Completion
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
				RawReferences: []model.CourseUserReference{"admin"},
			},
			"course-grader",
			"",
			RegradeResponse{
				RegradeResult: grader.RegradeResult{
					Results: map[string]*model.SubmissionHistoryItem{
						"course-admin@test.edulinq.org": nil,
					},
					WorkErrors: map[string]string{},
				},
				Complete:      true,
				ResolvedUsers: []string{"course-admin@test.edulinq.org"},
			},
		},

		// Student, No Wait
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: false,
				},
				RawReferences: []model.CourseUserReference{"student"},
			},
			"course-grader",
			"",
			RegradeResponse{
				RegradeResult: grader.RegradeResult{
					Results:    map[string]*model.SubmissionHistoryItem{},
					WorkErrors: map[string]string{},
				},
				Complete:      false,
				ResolvedUsers: []string{"course-student@test.edulinq.org"},
			},
		},

		// Grader, No Wait
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: false,
				},
				RawReferences: []model.CourseUserReference{"grader"},
			},
			"course-grader",
			"",
			RegradeResponse{
				RegradeResult: grader.RegradeResult{
					Results:    map[string]*model.SubmissionHistoryItem{},
					WorkErrors: map[string]string{},
				},
				Complete:      false,
				ResolvedUsers: []string{"course-grader@test.edulinq.org"},
			},
		},

		// Errors
		// Invalid Target Users
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
				RawReferences: []model.CourseUserReference{"ZZZ"},
			},
			"course-admin",
			"-638",
			RegradeResponse{},
		},
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: false,
				},
				RawReferences: []model.CourseUserReference{"ZZZ"},
			},
			"course-admin",
			"-638",
			RegradeResponse{},
		},

		// Perm errors
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
				RawReferences: []model.CourseUserReference{"student"},
			},
			"course-student",
			"-020",
			RegradeResponse{},
		},
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: false,
				},
				RawReferences: []model.CourseUserReference{"student"},
			},
			"course-other",
			"-020",
			RegradeResponse{},
		},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-users":        testCase.options.RawReferences,
			"wait-for-completion": testCase.options.JobOptions.WaitForCompletion,
		}

		var minRegradeAfter timestamp.Timestamp = 0
		if testCase.options.RegradeAfter == nil {
			// Create a window for the regrade after check.
			minRegradeAfter = timestamp.Now()
		} else {
			minRegradeAfter = *testCase.options.RegradeAfter
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/proxy/regrade`, fields, nil, testCase.proxyUser)
		if !response.Success {
			if testCase.expectedLocator != "" {
				if response.Locator != testCase.expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.expectedLocator != "" {
			test.Errorf("Case %d: Did not get an expected error.", i)
			continue
		}

		var responseContent RegradeResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		failed := grader.CheckAndClearIDs(test, i, testCase.expected.Results, responseContent.Results)
		if failed {
			continue
		}

		var maxRegradeAfter timestamp.Timestamp = 0
		if testCase.options.RegradeAfter == nil {
			// Create a window for the regrade after check.
			maxRegradeAfter = timestamp.Now()
		} else {
			maxRegradeAfter = *testCase.options.RegradeAfter
		}

		if !((minRegradeAfter <= responseContent.RegradeAfter) && (responseContent.RegradeAfter <= maxRegradeAfter)) {
			test.Errorf("Case %d: Unexpected regrade after time. Expected a time between '%d' and '%d', Actual: '%d'.",
				i, minRegradeAfter, maxRegradeAfter, responseContent.RegradeAfter)
			continue
		}

		// Clear regrade after time for equality check.
		responseContent.RegradeAfter = 0

		testCase.expected.Options = testCase.options
		if !reflect.DeepEqual(testCase.expected, responseContent) {
			test.Errorf("Case %d: Unexpected regrade result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}
	}
}
