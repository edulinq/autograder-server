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

func TestIndividualBase(test *testing.T) {
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

	response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/analysis/individual`, fields, nil, email)
	if !response.Success {
		test.Fatalf("Initial response is not a success when it should be: '%v'.", response)
	}

	var responseContent IndividualResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	// First round should have nothing, because we are not waiting for completion.
	expected := IndividualResponse{
		Complete: false,
		Options: analysis.AnalysisOptions{
			JobOptions:         jobmanager.JobOptions{},
			RawSubmissionSpecs: submissions,
		},
		Summary: &model.IndividualAnalysisSummary{
			AnalysisSummary: model.AnalysisSummary{
				Complete:      false,
				CompleteCount: 0,
				PendingCount:  2,
			},
		},
		Results: []*model.IndividualAnalysis{},
	}

	if !reflect.DeepEqual(expected, responseContent) {
		test.Fatalf("Initial response is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent))
	}

	// Make another request, but wait for the analysis.
	time.Sleep(100 * time.Millisecond)
	fields["wait-for-completion"] = true

	response = core.SendTestAPIRequestFull(test, `courses/assignments/submissions/analysis/individual`, fields, nil, email)
	if !response.Success {
		test.Fatalf("Second response is not a success when it should be: '%v'.", response)
	}

	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	// Second round should be complete.
	expected = IndividualResponse{
		Complete: true,
		Options: analysis.AnalysisOptions{
			RawSubmissionSpecs: submissions,
			JobOptions: jobmanager.JobOptions{
				WaitForCompletion: true,
			},
		},
		Summary: &model.IndividualAnalysisSummary{
			AnalysisSummary: model.AnalysisSummary{
				Complete:       true,
				CompleteCount:  2,
				PendingCount:   0,
				FirstTimestamp: timestamp.Zero(),
				LastTimestamp:  timestamp.Zero(),
			},
			AggregateScore: util.AggregateValues{
				Count:  2,
				Mean:   0.50,
				Median: 0.50,
				Min:    0,
				Max:    1,
			},
			AggregateLinesOfCode: util.AggregateValues{
				Count:  2,
				Mean:   4,
				Median: 4,
				Min:    4,
				Max:    4,
			},
			AggregateSubmissionTimeDelta: util.AggregateValues{
				Count:  2,
				Mean:   5000,
				Median: 5000,
				Min:    0,
				Max:    10000,
			},
			AggregateLinesOfCodeDelta: util.AggregateValues{
				Count:  2,
				Mean:   0,
				Median: 0,
				Min:    0,
				Max:    0,
			},
			AggregateScoreDelta: util.AggregateValues{
				Count:  2,
				Mean:   0.50,
				Median: 0.50,
				Min:    0,
				Max:    1,
			},
			AggregateLinesOfCodeVelocity: util.AggregateValues{
				Count:  2,
				Mean:   0,
				Median: 0,
				Min:    0,
				Max:    0,
			},
			AggregateScoreVelocity: util.AggregateValues{
				Count:  2,
				Mean:   180,
				Median: 180,
				Min:    0,
				Max:    360,
			},
			AggregateLinesOfCodePerFile: map[string]util.AggregateValues{
				"submission.py": util.AggregateValues{
					Count:  2,
					Mean:   4,
					Median: 4,
					Min:    4,
					Max:    4,
				},
			},
		},
		Results: []*model.IndividualAnalysis{
			&model.IndividualAnalysis{
				Options:             assignment.AssignmentAnalysisOptions,
				AnalysisTimestamp:   timestamp.Zero(),
				FullID:              "course101::hw0::course-student@test.edulinq.org::1697406256",
				ShortID:             "1697406256",
				CourseID:            "course101",
				AssignmentID:        "hw0",
				UserEmail:           "course-student@test.edulinq.org",
				SubmissionStartTime: timestamp.FromMSecs(1697406256000),
				Score:               0,
				LinesOfCode:         4,
				SubmissionTimeDelta: 0,
				LinesOfCodeDelta:    0,
				ScoreDelta:          0,
				LinesOfCodeVelocity: 0,
				ScoreVelocity:       0,
				Files: []model.AnalysisFileInfo{
					model.AnalysisFileInfo{
						Filename:    "submission.py",
						LinesOfCode: 4,
					},
				},
			},
			&model.IndividualAnalysis{
				Options:             assignment.AssignmentAnalysisOptions,
				AnalysisTimestamp:   timestamp.Zero(),
				FullID:              "course101::hw0::course-student@test.edulinq.org::1697406265",
				ShortID:             "1697406265",
				CourseID:            "course101",
				AssignmentID:        "hw0",
				UserEmail:           "course-student@test.edulinq.org",
				SubmissionStartTime: timestamp.FromMSecs(1697406266000),
				Score:               1,
				LinesOfCode:         4,
				SubmissionTimeDelta: 10000,
				LinesOfCodeDelta:    0,
				ScoreDelta:          1,
				LinesOfCodeVelocity: 0,
				ScoreVelocity:       360,
				Files: []model.AnalysisFileInfo{
					model.AnalysisFileInfo{
						Filename:    "submission.py",
						LinesOfCode: 4,
					},
				},
			},
		},
	}

	// Zero out the timestamps.
	responseContent.Summary.FirstTimestamp = timestamp.Zero()
	responseContent.Summary.LastTimestamp = timestamp.Zero()
	for _, result := range responseContent.Results {
		result.AnalysisTimestamp = timestamp.Zero()
	}

	// Normalize floats.
	for _, response := range []IndividualResponse{expected, responseContent} {
		response.Summary.RoundWithPrecision(2)
		for _, result := range response.Results {
			result.RoundWithPrecision(2)
		}
	}

	if !reflect.DeepEqual(expected, responseContent) {
		test.Fatalf("Second response is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent))
	}
}

func TestIndividualCheckPermissions(test *testing.T) {
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
