package db

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
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

		// Error Record
		{
			[]string{testIndividualRecords[2].FullID},
			map[string]*model.IndividualAnalysis{
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
			test.Errorf("Case %d: Results not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(results))
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
			test.Errorf("Case %d: Results not as expected. Expected: '%s', Actual: '%s'.",
				i, mustToJSONPairwiseMap(testCase.expected), mustToJSONPairwiseMap(results))
			continue
		}
	}
}

func (this *DBTests) DBTestRemoveIndividualAnalysisBase(test *testing.T) {
	defer ResetForTesting()

	testCases := []struct {
		removeIDs []string
		expected  map[string]*model.IndividualAnalysis
	}{
		{
			[]string{
				testIndividualIDs[0],
				testIndividualIDs[1],
				testIndividualIDs[2],
			},
			map[string]*model.IndividualAnalysis{},
		},
		{
			[]string{},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[0].FullID: testIndividualRecords[0],
				testIndividualRecords[1].FullID: testIndividualRecords[1],
				testIndividualRecords[2].FullID: testIndividualRecords[2],
			},
		},
		{
			[]string{
				"AAA",
			},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[0].FullID: testIndividualRecords[0],
				testIndividualRecords[1].FullID: testIndividualRecords[1],
				testIndividualRecords[2].FullID: testIndividualRecords[2],
			},
		},
		{
			[]string{
				testIndividualIDs[0],
			},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[1].FullID: testIndividualRecords[1],
				testIndividualRecords[2].FullID: testIndividualRecords[2],
			},
		},
		{
			[]string{
				testIndividualIDs[1],
			},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[0].FullID: testIndividualRecords[0],
				testIndividualRecords[2].FullID: testIndividualRecords[2],
			},
		},
		{
			[]string{
				testIndividualIDs[2],
			},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[0].FullID: testIndividualRecords[0],
				testIndividualRecords[1].FullID: testIndividualRecords[1],
			},
		},
		{
			[]string{
				testIndividualIDs[0],
				testIndividualIDs[1],
			},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[2].FullID: testIndividualRecords[2],
			},
		},
		{
			[]string{
				testIndividualIDs[0],
				testIndividualIDs[2],
			},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[1].FullID: testIndividualRecords[1],
			},
		},
		{
			[]string{
				testIndividualIDs[1],
				testIndividualIDs[2],
			},
			map[string]*model.IndividualAnalysis{
				testIndividualRecords[0].FullID: testIndividualRecords[0],
			},
		},
	}

	for i, testCase := range testCases {
		ResetForTesting()

		err := StoreIndividualAnalysis(testIndividualRecords)
		if err != nil {
			test.Errorf("Case %d: Failed to store initial records: '%v'.", i, err)
			continue
		}

		err = RemoveIndividualAnalysis(testCase.removeIDs)
		if err != nil {
			test.Errorf("Case %d: Failed to remove records: '%v'.", i, err)
			continue
		}

		results, err := GetIndividualAnalysis(testIndividualIDs)
		if err != nil {
			test.Errorf("Case %d: Failed to get records: '%v'.", i, err)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, results) {
			test.Errorf("Case %d: Results not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(results))
			continue
		}
	}
}

func (this *DBTests) DBTestRemovePairwiseAnalysisBase(test *testing.T) {
	defer ResetForTesting()

	testCases := []struct {
		removeIDs []model.PairwiseKey
		expected  map[model.PairwiseKey]*model.PairwiseAnalysis
	}{
		{
			[]model.PairwiseKey{
				testPairwiseKeys[0],
				testPairwiseKeys[1],
				testPairwiseKeys[2],
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{},
		},
		{
			[]model.PairwiseKey{},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
				testPairwiseRecords[1].SubmissionIDs: testPairwiseRecords[1],
				testPairwiseRecords[2].SubmissionIDs: testPairwiseRecords[2],
			},
		},
		{
			[]model.PairwiseKey{
				model.NewPairwiseKey("AAA", "BBB"),
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
				testPairwiseRecords[1].SubmissionIDs: testPairwiseRecords[1],
				testPairwiseRecords[2].SubmissionIDs: testPairwiseRecords[2],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseKeys[0],
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[1].SubmissionIDs: testPairwiseRecords[1],
				testPairwiseRecords[2].SubmissionIDs: testPairwiseRecords[2],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseKeys[1],
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
				testPairwiseRecords[2].SubmissionIDs: testPairwiseRecords[2],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseKeys[2],
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
				testPairwiseRecords[1].SubmissionIDs: testPairwiseRecords[1],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseKeys[0],
				testPairwiseKeys[1],
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[2].SubmissionIDs: testPairwiseRecords[2],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseKeys[0],
				testPairwiseKeys[2],
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[1].SubmissionIDs: testPairwiseRecords[1],
			},
		},
		{
			[]model.PairwiseKey{
				testPairwiseKeys[1],
				testPairwiseKeys[2],
			},
			map[model.PairwiseKey]*model.PairwiseAnalysis{
				testPairwiseRecords[0].SubmissionIDs: testPairwiseRecords[0],
			},
		},
	}

	for i, testCase := range testCases {
		ResetForTesting()

		err := StorePairwiseAnalysis(testPairwiseRecords)
		if err != nil {
			test.Errorf("Case %d: Failed to store initial records: '%v'.", i, err)
			continue
		}

		err = RemovePairwiseAnalysis(testCase.removeIDs)
		if err != nil {
			test.Errorf("Case %d: Failed to remove records: '%v'.", i, err)
			continue
		}

		results, err := GetPairwiseAnalysis(testPairwiseKeys)
		if err != nil {
			test.Errorf("Case %d: Failed to get records: '%v'.", i, err)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, results) {
			test.Errorf("Case %d: Results not as expected. Expected: '%s', Actual: '%s'.",
				i, mustToJSONPairwiseMap(testCase.expected), mustToJSONPairwiseMap(results))
			continue
		}
	}
}

var testIndividualIDs []string = []string{
	"course101::hw0::course-student@test.edulinq.org::1697406256",
	"course101::hw0::course-student@test.edulinq.org::1697406265",
	"course101::hw0::course-student@test.edulinq.org::1697406272",
}

var testIndividualRecords []*model.IndividualAnalysis = []*model.IndividualAnalysis{
	&model.IndividualAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		FullID:            "course101::hw0::course-student@test.edulinq.org::1697406256",
		CourseID:          "course101",
	},
	&model.IndividualAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		FullID:            "course101::hw0::course-student@test.edulinq.org::1697406265",
		CourseID:          "course101",
	},
	&model.IndividualAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		FullID:            "course101::hw0::course-student@test.edulinq.org::1697406272",
		CourseID:          "course101",
		Failure:           true,
		FailureMessage:    "Analysis failed.",
	},
}

var testPairwiseKeys []model.PairwiseKey = []model.PairwiseKey{
	model.NewPairwiseKey(
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406265",
	),
	model.NewPairwiseKey(
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406272",
	),
	model.NewPairwiseKey(
		"course101::hw0::course-student@test.edulinq.org::1697406265",
		"course101::hw0::course-student@test.edulinq.org::1697406272",
	),
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
	},
	&model.PairwiseAnalysis{
		AnalysisTimestamp: timestamp.Zero(),
		SubmissionIDs: model.NewPairwiseKey(
			"course101::hw0::course-student@test.edulinq.org::1697406265",
			"course101::hw0::course-student@test.edulinq.org::1697406272",
		),
		Failure:        true,
		FailureMessage: "Analysis failed.",
	},
}

func mustToJSONPairwiseMap(input map[model.PairwiseKey]*model.PairwiseAnalysis) string {
	output := make(map[string]*model.PairwiseAnalysis, len(input))

	for key, value := range input {
		output[key.String()] = value
	}

	return util.MustToJSONIndent(output)
}
