package db

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

func (this *DBTests) DBTestGetIndividualAnalysisBase(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	err := StoreIndividualAnalysis(testIndividualRecords)
	if err != nil {
		test.Fatalf("Failed to store initial records: '%v'.", err)
	}

	testCases := []struct {
		fullSubmissionIDs []string
		expected          map[string]*model.IndividualAnalysis
	}{
		{
			[]string{},
			map[string]*model.IndividualAnalysis{},
		},
		{
			[]string{"A"},
			map[string]*model.IndividualAnalysis{},
		},
		{
			[]string{testIndividualRecords[0].FullID},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[0].FullID: testIndividualRecords[0],
			},
		},
		{
			[]string{
				testIndividualRecords[0].FullID,
				"A",
			},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[0].FullID: testIndividualRecords[0],
			},
		},
		{
			[]string{
				testIndividualRecords[0].FullID,
				testIndividualRecords[2].FullID,
			},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[0].FullID: testIndividualRecords[0],
				testIndividualRecords[2].FullID: testIndividualRecords[2],
			},
		},
	}

	for i, testCase := range testCases {
		results, err := GetIndividualAnalysis(testCase.fullSubmissionIDs)
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

func (this *DBTests) DBTestGetPairwiseAnalysisBase(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	err := StorePairwiseAnalysis(testPairwiseRecords)
	if err != nil {
		test.Fatalf("Failed to store initial records: '%v'.", err)
	}

	testCases := []struct {
		keys     []model.PairwiseKey
		expected map[model.PairwiseKey]*model.PairwiseAnalysis
	}{
		{
			[]model.PairwiseKey{},
			map[model.PairwiseKey]*model.PairwiseAnalysis{},
		},
		{
			[]model.PairwiseKey{model.NewPairwiseKey("A", "Z")},
			map[model.PairwiseKey]*model.PairwiseAnalysis{},
		},
		{
			[]model.PairwiseKey{testPairwiseRecords[0].SubmissionIDs},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseRecords[0].SubmissionIDs,
				model.NewPairwiseKey("A", "Z"),
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseRecords[0].SubmissionIDs,
				testPairwiseRecords[2].SubmissionIDs,
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
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

var testIndividualRecords []*model.IndividualAnalysis = []*model.IndividualAnalysis{
	&model.IndividualAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		FullID:            "course101::hw0::course-student@test.edulinq.org::1697406256",
	},
	&model.IndividualAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		FullID:            "course101::hw0::course-student@test.edulinq.org::1697406265",
	},
	&model.IndividualAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		FullID:            "course101::hw0::course-student@test.edulinq.org::1697406272",
	},
}

var testPairwiseRecords []*model.PairwiseAnalysis = []*model.PairwiseAnalysis{
	&model.PairwiseAnalysis{
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
					Version:  "0.0.1",
					Score:    0.13,
				},
			},
		},
		UnmatchedFiles: [][2]string{},
	},
	&model.PairwiseAnalysis{
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
					Version:  "0.0.1",
					Score:    0.13,
				},
			},
		},
		UnmatchedFiles: [][2]string{},
	},
	&model.PairwiseAnalysis{
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
					Version:  "0.0.1",
					Score:    0.13,
				},
			},
		},
		UnmatchedFiles: [][2]string{},
	},
}
