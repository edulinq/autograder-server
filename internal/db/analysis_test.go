package db

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

func (this *DBTests) DBTestGetPairwiseAnalysisBase(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	err := StorePairwiseAnalysis(testPairwiseRecords)
	if err != nil {
		test.Fatalf("Failed to store initial records: '%v'.", err)
	}

	testCases := []struct {
		keys     []model.PairwiseKey
		expected map[model.PairwiseKey]*model.PairWiseAnalysis
	}{
		{
			[]model.PairwiseKey{},
			map[model.PairwiseKey]*model.PairWiseAnalysis{},
		},
		{
			[]model.PairwiseKey{model.NewPairwiseKey("A", "Z")},
			map[model.PairwiseKey]*model.PairWiseAnalysis{},
		},
		{
			[]model.PairwiseKey{testPairwiseRecords[0].SubmissionIDs},
			map[model.PairwiseKey]*model.PairWiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseRecords[0].SubmissionIDs,
				model.NewPairwiseKey("A", "Z"),
			},
			map[model.PairwiseKey]*model.PairWiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseRecords[0].SubmissionIDs,
				testPairwiseRecords[2].SubmissionIDs,
			},
			map[model.PairwiseKey]*model.PairWiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
				testPairwiseRecords[2].SubmissionIDs: testPairwiseRecords[2],
			},
		},
	}

	for i, testCase := range testCases {
		results, err := GetPairwiseAnalysis(testCase.keys)
		if err != nil {
			test.Errorf("Case %d: Failed to get records: '%v'.", i, err)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, results) {
			test.Errorf("Case %d: Results not as expected. Expected: '%v', Actual: '%v'.",
				i, testCase.expected, results)
			continue
		}
	}
}

var testPairwiseRecords []*model.PairWiseAnalysis = []*model.PairWiseAnalysis{
	&model.PairWiseAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		SubmissionIDs: model.NewPairwiseKey(
			"course101::hw0::course-student@test.edulinq.org::1697406256",
			"course101::hw0::course-student@test.edulinq.org::1697406265",
		),
		Similarities: map[string][]*model.FileSimilarity{
			"submission.py": []*model.FileSimilarity{
				&model.FileSimilarity{
					Filename: "submission.py",
					Tool:     "fake",
					Score:    0.13,
				},
			},
		},
		UnmatchedFiles: [][2]string{},
	},
	&model.PairWiseAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		SubmissionIDs: model.NewPairwiseKey(
			"course101::hw0::course-student@test.edulinq.org::1697406256",
			"course101::hw0::course-student@test.edulinq.org::1697406272",
		),
		Similarities: map[string][]*model.FileSimilarity{
			"submission.py": []*model.FileSimilarity{
				&model.FileSimilarity{
					Filename: "submission.py",
					Tool:     "fake",
					Score:    0.13,
				},
			},
		},
		UnmatchedFiles: [][2]string{},
	},
	&model.PairWiseAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		SubmissionIDs: model.NewPairwiseKey(
			"course101::hw0::course-student@test.edulinq.org::1697406265",
			"course101::hw0::course-student@test.edulinq.org::1697406272",
		),
		Similarities: map[string][]*model.FileSimilarity{
			"submission.py": []*model.FileSimilarity{
				&model.FileSimilarity{
					Filename: "submission.py",
					Tool:     "fake",
					Score:    0.13,
				},
			},
		},
		UnmatchedFiles: [][2]string{},
	},
}
