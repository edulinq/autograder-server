package proxy

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestRegradeBase(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	// Note that computation of these paths are deferred until test time.
	studentGradingResults := map[string]*model.GradingResult{
		"1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
	}

	testCases := []struct {
		role              string
		proxyUser         string
		waitForCompletion bool
		expectedLocator   string
		expected          RegradeResponse
	}{
		// Valid regrade submissions
		// Student, wait for completion
		{
			"student",
			"course-grader",
			true,
			"",
			RegradeResponse{
				Complete: true,
				Users:    []string{"course-student@test.edulinq.org"},
				Results: map[string]*model.SubmissionHistoryItem{
					"course-student@test.edulinq.org": studentGradingResults["1697406272"].Info.ToHistoryItem(),
				},
			},
		},

		// Admin, wait for completion
		{
			"admin",
			"course-grader",
			true,
			"",
			RegradeResponse{
				Complete: true,
				Users:    []string{"course-admin@test.edulinq.org"},
				Results: map[string]*model.SubmissionHistoryItem{
					"course-admin@test.edulinq.org": nil,
				},
			},
		},

		// Student, no wait
		{
			"student",
			"course-grader",
			false,
			"",
			RegradeResponse{
				Complete: false,
				Users:    []string{"course-student@test.edulinq.org"},
				Results:  map[string]*model.SubmissionHistoryItem{},
			},
		},

		// Grader, no wait
		{
			"grader",
			"course-grader",
			false,
			"",
			RegradeResponse{
				Complete: false,
				Users:    []string{"course-grader@test.edulinq.org"},
				Results:  map[string]*model.SubmissionHistoryItem{},
			},
		},

		// Invalid regrade submissions
		// Unknown role, wait
		{
			"ZZZ",
			"course-admin",
			true,
			"-005",
			RegradeResponse{},
		},
		{
			"",
			"course-admin",
			true,
			"-005",
			RegradeResponse{},
		},

		// Unknown role, no wait
		{
			"ZZZ",
			"course-admin",
			false,
			"-005",
			RegradeResponse{},
		},

		// Perm errors
		{
			"student",
			"course-student",
			false,
			"-020",
			RegradeResponse{},
		},
		{
			"student",
			"course-other",
			true,
			"-020",
			RegradeResponse{},
		},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"filter-role":         testCase.role,
			"wait-for-completion": testCase.waitForCompletion,
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
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		var responseContent RegradeResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		failed := grader.CheckAndClearIDs(test, i, testCase.expected.Results, responseContent.Results)
		if failed {
			continue
		}

		if !reflect.DeepEqual(testCase.expected, responseContent) {
			test.Errorf("Case %d: Unexpected regrade result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}
	}
}
