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

	// A time in the year 3003.
	farFutureTime := timestamp.FromMSecs(32614181465000)

	testCases := []struct {
		options         grader.RegradeOptions
		proxyUser       string
		expectedLocator string
		expected        RegradeResponse
	}{
		// Note: Tests that do not wait for completion are left out because of flakiness.

		// Student
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

		// Empty Users
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
			},
			"course-grader",
			"",
			RegradeResponse{
				RegradeResult: grader.RegradeResult{
					Results: map[string]*model.SubmissionHistoryItem{
						"course-admin@test.edulinq.org":   nil,
						"course-grader@test.edulinq.org":  nil,
						"course-other@test.edulinq.org":   nil,
						"course-owner@test.edulinq.org":   nil,
						"course-student@test.edulinq.org": studentGradingResults["1697406272"].Info.ToHistoryItem(),
					},
					WorkErrors: map[string]string{},
				},
				Complete: true,
				ResolvedUsers: []string{
					"course-admin@test.edulinq.org",
					"course-grader@test.edulinq.org",
					"course-other@test.edulinq.org",
					"course-owner@test.edulinq.org",
					"course-student@test.edulinq.org",
				},
			},
		},

		// All Users
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
				RawReferences: []model.CourseUserReference{"*"},
			},
			"course-grader",
			"",
			RegradeResponse{
				RegradeResult: grader.RegradeResult{
					Results: map[string]*model.SubmissionHistoryItem{
						"course-admin@test.edulinq.org":   nil,
						"course-grader@test.edulinq.org":  nil,
						"course-other@test.edulinq.org":   nil,
						"course-owner@test.edulinq.org":   nil,
						"course-student@test.edulinq.org": studentGradingResults["1697406272"].Info.ToHistoryItem(),
					},
					WorkErrors: map[string]string{},
				},
				Complete: true,
				ResolvedUsers: []string{
					"course-admin@test.edulinq.org",
					"course-grader@test.edulinq.org",
					"course-other@test.edulinq.org",
					"course-owner@test.edulinq.org",
					"course-student@test.edulinq.org",
				},
			},
		},

		// Early Regrade After
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
				RawReferences: []model.CourseUserReference{"student"},
				RegradeCutoff: timestamp.ZeroPointer(),
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

		// Late Regrade After
		{
			grader.RegradeOptions{
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
				RawReferences: []model.CourseUserReference{"student"},
				RegradeCutoff: &farFutureTime,
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

		// Perm Errors
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
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-users":        testCase.options.RawReferences,
			"wait-for-completion": testCase.options.JobOptions.WaitForCompletion,
			"regrade-cutoff":      testCase.options.RegradeCutoff,
		}

		// Update empty raw references to be the "*" to pass equality check.
		if len(testCase.options.RawReferences) == 0 {
			testCase.options.RawReferences = model.NewAllCourseUserReference()
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/proxy/regrade`, fields, nil, testCase.proxyUser)
		if !response.Success {
			if testCase.expectedLocator != "" {
				if response.Locator != testCase.expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected: '%s', Actual: '%s'.",
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

		// Copy the test case options to the expected output to pass the equality check.
		testCase.expected.Options = testCase.options

		// Clear variable regrade after times for equality check.
		if testCase.options.RegradeCutoff == nil {
			testCase.expected.Options.RegradeCutoff = responseContent.Options.RegradeCutoff
		}

		// Clear the submission IDs to pass the equality check.
		for _, expected := range testCase.expected.Results {
			if expected == nil {
				continue
			}

			expected.ShortID = ""
			expected.ID = ""
		}

		for _, actual := range responseContent.Results {
			if actual == nil {
				continue
			}

			actual.ShortID = ""
			actual.ID = ""
		}

		if !reflect.DeepEqual(testCase.expected, responseContent) {
			test.Errorf("Case %d: Unexpected regrade result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}
	}
}
