package model

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestAssignmentAnalysisOptionsValidateBase(test *testing.T) {
	testCases := []struct {
		input          *AssignmentAnalysisOptions
		expected       *AssignmentAnalysisOptions
		errorSubstring string
	}{
		{
			&AssignmentAnalysisOptions{},
			&AssignmentAnalysisOptions{
				IncludePatterns: []string{
					DEFAULT_INCLUDE_REGEX,
				},
			},
			"",
		},
		{
			&AssignmentAnalysisOptions{
				IncludePatterns: []string{
					"AAA",
				},
				ExcludePatterns: []string{
					"ZZZ",
				},
			},
			&AssignmentAnalysisOptions{
				IncludePatterns: []string{
					"AAA",
				},
				ExcludePatterns: []string{
					"ZZZ",
				},
			},
			"",
		},
		{
			&AssignmentAnalysisOptions{
				TemplateFiles: []*util.FileSpec{
					&util.FileSpec{
						Type: util.FILESPEC_TYPE_PATH,
						Path: " a/../b ",
					},
				},
				TemplateFileOps: []*util.FileOperation{
					util.NewFileOperation([]string{"copy", " a/../b ", "c"}),
				},
			},
			&AssignmentAnalysisOptions{
				IncludePatterns: []string{
					DEFAULT_INCLUDE_REGEX,
				},
				TemplateFiles: []*util.FileSpec{
					&util.FileSpec{
						Type: util.FILESPEC_TYPE_PATH,
						Path: "b",
					},
				},
				TemplateFileOps: []*util.FileOperation{
					util.NewFileOperation([]string{"copy", "b", "c"}),
				},
			},
			"",
		},

		// Errors

		// Nil
		{
			nil,
			nil,
			"cannot be nil",
		},
		// Bad Include/Exclude Patterns
		{
			&AssignmentAnalysisOptions{
				IncludePatterns: []string{
					"(error",
				},
			},
			nil,
			"Failed to compile include pattern",
		},
		{
			&AssignmentAnalysisOptions{
				ExcludePatterns: []string{
					"(error",
				},
			},
			nil,
			"Failed to compile exclude pattern",
		},
		// Bad Template Paths
		{
			&AssignmentAnalysisOptions{
				TemplateFiles: []*util.FileSpec{
					&util.FileSpec{
						Type: util.FILESPEC_TYPE_PATH,
						Path: "/a",
					},
				},
			},
			nil,
			"not allowed to be absolute",
		},
		{
			&AssignmentAnalysisOptions{
				TemplateFiles: []*util.FileSpec{
					&util.FileSpec{
						Type: util.FILESPEC_TYPE_PATH,
						Path: "../a",
					},
				},
			},
			nil,
			"outside of the its base directory",
		},
		{
			&AssignmentAnalysisOptions{
				TemplateFiles: []*util.FileSpec{
					&util.FileSpec{
						Type: util.FILESPEC_TYPE_URL,
						Path: "http://test.edulinq.org/a.zip",
						Dest: "../a.zip",
					},
				},
			},
			nil,
			"outside of the its base directory",
		},
		{
			&AssignmentAnalysisOptions{
				TemplateFiles: []*util.FileSpec{
					&util.FileSpec{
						Type: util.FILESPEC_TYPE_GIT,
						Path: "http://test.edulinq.org/a.git",
						Dest: "../a",
					},
				},
			},
			nil,
			"outside of the its base directory",
		},
		{
			&AssignmentAnalysisOptions{
				TemplateFileOps: []*util.FileOperation{
					util.NewFileOperation([]string{"copy", "/a", "b"}),
				},
			},
			nil,
			"Only relative paths are allowed",
		},
		{
			&AssignmentAnalysisOptions{
				TemplateFileOps: []*util.FileOperation{
					util.NewFileOperation([]string{"copy", "../a", "b"}),
				},
			},
			nil,
			"outside of the its base directory",
		},
	}

	for i, testCase := range testCases {
		err := testCase.input.Validate()
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to validate '%+v': '%v'.", i, testCase.input, err)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error: '%s'.", i, testCase.errorSubstring)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, testCase.input) {
			test.Errorf("Case %d: Result not as expected. Expected: '%+v', Actual: '%+v'.",
				i, testCase.expected, testCase.input)
			continue
		}
	}
}

func TestAssignmentAnalysisOptionsMatchRelpathBase(test *testing.T) {
	testCases := []struct {
		options  *AssignmentAnalysisOptions
		relpath  string
		expected bool
	}{
		// Default values.
		{
			&AssignmentAnalysisOptions{},
			"ZZZ",
			true,
		},
		{
			&AssignmentAnalysisOptions{},
			"",
			false,
		},

		// (include && exclude).
		{
			&AssignmentAnalysisOptions{
				IncludePatterns: []string{
					"A",
				},
				ExcludePatterns: []string{
					"B",
				},
			},
			"AB",
			false,
		},

		// (include && !exclude).
		{
			&AssignmentAnalysisOptions{
				IncludePatterns: []string{
					"A",
				},
				ExcludePatterns: []string{
					"B",
				},
			},
			"AC",
			true,
		},

		// (!include && exclude).
		{
			&AssignmentAnalysisOptions{
				IncludePatterns: []string{
					"A",
				},
				ExcludePatterns: []string{
					"B",
				},
			},
			"B",
			false,
		},

		// (!include && !exclude).
		{
			&AssignmentAnalysisOptions{
				IncludePatterns: []string{
					"A",
				},
				ExcludePatterns: []string{
					"B",
				},
			},
			"Z",
			false,
		},
	}

	for i, testCase := range testCases {
		err := testCase.options.Validate()
		if err != nil {
			test.Errorf("Case %d: Options do not validate: '%v'.", i, err)
			continue
		}

		actual := testCase.options.MatchRelpath(testCase.relpath)
		if testCase.expected != actual {
			test.Errorf("Case %d: Result not as expected. Expected: '%v', Actual: '%v'.",
				i, testCase.expected, actual)
			continue
		}
	}
}

func TestNewIndividualAnalysisSummaryBase(test *testing.T) {
	input := map[string]*IndividualAnalysis{
		"A": &IndividualAnalysis{
			Score:               10,
			LinesOfCode:         10,
			SubmissionTimeDelta: 0,
			LinesOfCodeDelta:    0,
			ScoreDelta:          0,
			LinesOfCodeVelocity: 0,
			ScoreVelocity:       0,
			Files: []AnalysisFileInfo{
				AnalysisFileInfo{
					Filename:    "a.go",
					LinesOfCode: 10,
				},
			},
		},
		"B": &IndividualAnalysis{
			Score:               20,
			LinesOfCode:         40,
			SubmissionTimeDelta: 12,
			LinesOfCodeDelta:    15,
			ScoreDelta:          20,
			LinesOfCodeVelocity: 25,
			ScoreVelocity:       30,
			Files: []AnalysisFileInfo{
				AnalysisFileInfo{
					Filename:    "a.go",
					LinesOfCode: 20,
				},
				AnalysisFileInfo{
					Filename:    "b.go",
					LinesOfCode: 20,
				},
			},
		},
		"C": &IndividualAnalysis{
			Score:               30,
			LinesOfCode:         20,
			SubmissionTimeDelta: 32,
			LinesOfCodeDelta:    35,
			ScoreDelta:          40,
			LinesOfCodeVelocity: 45,
			ScoreVelocity:       50,
			Files: []AnalysisFileInfo{
				AnalysisFileInfo{
					Filename:    "a.go",
					LinesOfCode: 20,
				},
			},
		},
		"FAILURE": &IndividualAnalysis{
			Failure: true,
		},
	}

	expected := &IndividualAnalysisSummary{
		AnalysisSummary: AnalysisSummary{
			Complete:       true,
			CompleteCount:  3,
			PendingCount:   0,
			FailureCount:   1,
			FirstTimestamp: timestamp.Zero(),
			LastTimestamp:  timestamp.Zero(),
		},
		AggregateScore: util.AggregateValues{
			Count:  3,
			Mean:   20,
			Median: 20,
			Min:    10,
			Max:    30,
		},
		AggregateLinesOfCode: util.AggregateValues{
			Count:  3,
			Mean:   23.33,
			Median: 20,
			Min:    10,
			Max:    40,
		},
		AggregateSubmissionTimeDelta: util.AggregateValues{
			Count:  3,
			Mean:   14.67,
			Median: 12,
			Min:    0,
			Max:    32,
		},
		AggregateLinesOfCodeDelta: util.AggregateValues{
			Count:  3,
			Mean:   16.67,
			Median: 15,
			Min:    0,
			Max:    35,
		},
		AggregateScoreDelta: util.AggregateValues{
			Count:  3,
			Mean:   20,
			Median: 20,
			Min:    0,
			Max:    40,
		},
		AggregateLinesOfCodeVelocity: util.AggregateValues{
			Count:  3,
			Mean:   23.33,
			Median: 25,
			Min:    0,
			Max:    45,
		},
		AggregateScoreVelocity: util.AggregateValues{
			Count:  3,
			Mean:   26.67,
			Median: 30,
			Min:    0,
			Max:    50,
		},
		AggregateLinesOfCodePerFile: map[string]util.AggregateValues{
			"a.go": util.AggregateValues{
				Count:  3,
				Mean:   16.67,
				Median: 20,
				Min:    10,
				Max:    20,
			},
			"b.go": util.AggregateValues{
				Count:  1,
				Mean:   20,
				Median: 20,
				Min:    20,
				Max:    20,
			},
		},
	}

	actual := NewIndividualAnalysisSummary(input, 0)

	// Zero out timesstamps.
	actual.FirstTimestamp = timestamp.Zero()
	actual.LastTimestamp = timestamp.Zero()

	// Normalize values.
	expected.RoundWithPrecision(2)
	actual.RoundWithPrecision(2)

	if !reflect.DeepEqual(expected, actual) {
		test.Fatalf("Incorrect result. Expected: '%s', Actual: '%s'.", util.MustToJSONIndent(expected), util.MustToJSONIndent(actual))
	}
}

func TestNewPairwiseAnalysisSummaryBase(test *testing.T) {
	sims1 := map[string][]*FileSimilarity{
		"a.py": []*FileSimilarity{
			&FileSimilarity{
				Filename:         "a.py",
				OriginalFilename: "a.ipynb",
				Tool:             "1",
				Score:            0.10,
			},
			&FileSimilarity{
				Filename: "a.py",
				Tool:     "2",
				Score:    0.20,
			},
		},
		"b.py": []*FileSimilarity{
			&FileSimilarity{
				Filename: "b.py",
				Tool:     "1",
				Score:    0.30,
			},
			&FileSimilarity{
				Filename: "b.py",
				Tool:     "2",
				Score:    0.40,
			},
		},
	}

	sims2 := map[string][]*FileSimilarity{
		"b.py": []*FileSimilarity{
			&FileSimilarity{
				Filename: "b.py",
				Tool:     "1",
				Score:    0.50,
			},
		},
		"c.py": []*FileSimilarity{
			&FileSimilarity{
				Filename: "c.py",
				Tool:     "2",
				Score:    0.60,
			},
		},
	}

	sims3 := map[string][]*FileSimilarity{
		"b.py": []*FileSimilarity{
			&FileSimilarity{
				Filename: "b.py",
				Tool:     "2",
				Score:    0.70,
			},
		},
		"c.py": []*FileSimilarity{
			&FileSimilarity{
				Filename: "c.py",
				Tool:     "3",
				Score:    0.80,
			},
		},
	}

	input := PairwiseAnalysisMap{
		NewPairwiseKey("A", "B"): NewPairwiseAnalysis(NewPairwiseKey("A", "B"), nil, sims1, nil, nil),
		NewPairwiseKey("C", "D"): NewPairwiseAnalysis(NewPairwiseKey("C", "D"), nil, sims2, nil, nil),
		NewPairwiseKey("E", "F"): NewPairwiseAnalysis(NewPairwiseKey("E", "F"), nil, sims3, nil, nil),
		NewPairwiseKey("FOO", "BAR"): &PairwiseAnalysis{
			Failure: true,
		},
	}

	expected := &PairwiseAnalysisSummary{
		AnalysisSummary: AnalysisSummary{
			Complete:       true,
			CompleteCount:  3,
			PendingCount:   0,
			FailureCount:   1,
			FirstTimestamp: timestamp.Zero(),
			LastTimestamp:  timestamp.Zero(),
		},
		AggregateMeanSimilarities: map[string]util.AggregateValues{
			"a.py": util.AggregateValues{
				Count:  1,
				Mean:   0.15,
				Median: 0.15,
				Min:    0.15,
				Max:    0.15,
			},
			"b.py": util.AggregateValues{
				Count:  3,
				Mean:   0.51666666,
				Median: 0.50,
				Min:    0.35,
				Max:    0.70,
			},
			"c.py": util.AggregateValues{
				Count:  2,
				Mean:   0.70,
				Median: 0.70,
				Min:    0.60,
				Max:    0.80,
			},
		},
		AggregateTotalMeanSimilarities: util.AggregateValues{
			Count:  3,
			Mean:   0.51666666,
			Median: 0.55,
			Min:    0.25,
			Max:    0.75,
		},
	}

	actual := NewPairwiseAnalysisSummary(input, 0)

	// Zero out timesstamps.
	actual.FirstTimestamp = timestamp.Zero()
	actual.LastTimestamp = timestamp.Zero()

	// Normalize values.
	expected.RoundWithPrecision(2)
	actual.RoundWithPrecision(2)

	if !reflect.DeepEqual(expected, actual) {
		test.Fatalf("Incorrect result. Expected: '%s', Actual: '%s'.", util.MustToJSONIndent(expected), util.MustToJSONIndent(actual))
	}
}

func TestAssignmentAnalysisOptionsFetchTemplateFilesBase(test *testing.T) {
	testCases := []struct {
		options          *AssignmentAnalysisOptions
		expectedRelpaths []string
		errorSubstring   string
	}{
		{
			&AssignmentAnalysisOptions{},
			[]string{},
			"",
		},
		{
			&AssignmentAnalysisOptions{
				TemplateFiles: []*util.FileSpec{
					&util.FileSpec{
						Type: util.FILESPEC_TYPE_PATH,
						Path: "a.txt",
					},
				},
			},
			[]string{
				"a.txt",
			},
			"",
		},
		{
			&AssignmentAnalysisOptions{
				TemplateFiles: []*util.FileSpec{
					&util.FileSpec{
						Type: util.FILESPEC_TYPE_PATH,
						Path: "a.txt",
					},
				},
				TemplateFileOps: []*util.FileOperation{
					util.NewFileOperation([]string{"copy", "a.txt", "b/c.txt"}),
				},
			},
			[]string{
				"a.txt",
				filepath.Join("b", "c.txt"),
			},
			"",
		},
		{
			&AssignmentAnalysisOptions{
				TemplateFiles: []*util.FileSpec{
					&util.FileSpec{
						Type: util.FILESPEC_TYPE_PATH,
						Path: "a.txt",
					},
				},
				TemplateFileOps: []*util.FileOperation{
					util.NewFileOperation([]string{"move", "a.txt", "b/c.txt"}),
				},
			},
			[]string{
				filepath.Join("b", "c.txt"),
			},
			"",
		},
	}

	// The base for the file specs.
	baseDir := filepath.Join(util.TestdataDirForTesting(), "files")

	for i, testCase := range testCases {
		err := testCase.options.Validate()
		if err != nil {
			test.Errorf("Case %d: Failed to validate options: '%v'.", i, err)
			continue
		}

		tempDir := util.MustMkDirTemp("test-analysis-fetch-templates-")
		defer util.RemoveDirent(tempDir)

		actualRelpaths, err := testCase.options.FetchTemplateFiles(baseDir, baseDir, tempDir)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to fetch template files '%+v': '%v'.", i, testCase.options, err)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error: '%s'.", i, testCase.errorSubstring)
			continue
		}

		if !reflect.DeepEqual(testCase.expectedRelpaths, actualRelpaths) {
			test.Errorf("Case %d: Result relpaths not as expected. Expected: '%v', Actual: '%v'.",
				i, testCase.expectedRelpaths, actualRelpaths)
			continue
		}
	}
}
