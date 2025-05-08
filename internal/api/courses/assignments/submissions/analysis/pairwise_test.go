package analysis

import (
	"reflect"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/analysis"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestPairwiseBase(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestAssignment()

	// Make an initial request, but don't wait.

	submissions := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406265",
	}

	email := "server-admin"
	fields := map[string]any{
		"submissions":         submissions,
		"wait-for-completion": false,
	}

	response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/analysis/pairwise`, fields, nil, email)
	if !response.Success {
		test.Fatalf("Initial response is not a success when it should be: '%v'.", response)
	}

	var responseContent PairwiseResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	// First round should have nothing, because we are not waiting for completion.
	expected := PairwiseResponse{
		Complete: false,
		Options: analysis.AnalysisOptions{
			JobOptions:         jobmanager.JobOptions{},
			RawSubmissionSpecs: submissions,
		},
		Summary: &model.PairwiseAnalysisSummary{
			AnalysisSummary: model.AnalysisSummary{
				Complete:      false,
				CompleteCount: 0,
				PendingCount:  1,
			},
		},
		Results: model.PairwiseAnalysisMap{},
	}

	if !reflect.DeepEqual(expected, responseContent) {
		test.Fatalf("Initial response is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent))
	}

	// Make another request, but wait for the analysis.
	time.Sleep(100 * time.Millisecond)
	fields["wait-for-completion"] = true

	response = core.SendTestAPIRequestFull(test, `courses/assignments/submissions/analysis/pairwise`, fields, nil, email)
	if !response.Success {
		test.Fatalf("Second response is not a success when it should be: '%v'.", response)
	}

	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	submissionID1 := "course101::hw0::course-student@test.edulinq.org::1697406256"
	submissionID2 := "course101::hw0::course-student@test.edulinq.org::1697406265"

	// Second round should be complete.
	expected = PairwiseResponse{
		Complete: true,
		Options: analysis.AnalysisOptions{
			RawSubmissionSpecs: submissions,
			JobOptions: jobmanager.JobOptions{
				WaitForCompletion: true,
			},
		},
		Summary: &model.PairwiseAnalysisSummary{
			AnalysisSummary: model.AnalysisSummary{
				Complete:       true,
				CompleteCount:  1,
				PendingCount:   0,
				FirstTimestamp: timestamp.Zero(),
				LastTimestamp:  timestamp.Zero(),
			},
			AggregateMeanSimilarities: map[string]util.AggregateValues{
				"submission.py": util.AggregateValues{
					Count:  1,
					Mean:   0.13,
					Median: 0.13,
					Min:    0.13,
					Max:    0.13,
				},
			},
			AggregateTotalMeanSimilarities: util.AggregateValues{
				Count:  1,
				Mean:   0.13,
				Median: 0.13,
				Min:    0.13,
				Max:    0.13,
			},
		},
		Results: model.PairwiseAnalysisMap{
			model.NewPairwiseKey(submissionID1, submissionID2): &model.PairwiseAnalysis{
				Options:           assignment.AssignmentAnalysisOptions,
				AnalysisTimestamp: timestamp.Zero(),
				SubmissionIDs: model.NewPairwiseKey(
					submissionID1,
					submissionID2,
				),
				Similarities: map[string][]*model.FileSimilarity{
					"submission.py": []*model.FileSimilarity{
						&model.FileSimilarity{
							Filename: "submission.py",
							Tool:     "fake",
							Version:  "0.0.1",
							Score:    0.13,
						},
					},
				},
				MeanSimilarities: map[string]float64{
					"submission.py": 0.13,
				},
				TotalMeanSimilarity: 0.13,
			},
		},
	}

	// Zero out the timestamps.
	responseContent.Summary.FirstTimestamp = timestamp.Zero()
	responseContent.Summary.LastTimestamp = timestamp.Zero()
	for _, result := range responseContent.Results {
		result.AnalysisTimestamp = timestamp.Zero()
	}

	if !reflect.DeepEqual(expected, responseContent) {
		test.Fatalf("Second response is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent))
	}
}

func TestPairwiseCheckPermissions(test *testing.T) {
	testCases := []struct {
		user     *model.ServerUser
		courses  []string
		expected bool
	}{
		{
			db.MustGetServerUser("server-admin@test.edulinq.org"),
			[]string{"course101"},
			true,
		},
		{
			db.MustGetServerUser("server-user@test.edulinq.org"),
			[]string{"course101"},
			false,
		},
		{
			db.MustGetServerUser("course-grader@test.edulinq.org"),
			[]string{"course101"},
			true,
		},
		{
			db.MustGetServerUser("course-student@test.edulinq.org"),
			[]string{"course101"},
			false,
		},
		{
			db.MustGetServerUser("course-grader@test.edulinq.org"),
			[]string{"course101", "course-languages"},
			true,
		},
		{
			db.MustGetServerUser("course-student@test.edulinq.org"),
			[]string{"course101", "course-languages"},
			false,
		},
		{
			db.MustGetServerUser("server-admin@test.edulinq.org"),
			[]string{"course101", "course-languages"},
			true,
		},
	}

	for i, testCase := range testCases {
		actual := checkPermissions(testCase.user, testCase.courses)
		if testCase.expected != actual {
			test.Errorf("Case %d: Incorrect. Expected: %v, Actual: %v.", i, testCase.expected, actual)
		}
	}
}
