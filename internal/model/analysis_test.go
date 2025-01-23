package model

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestNewIndividualAnalysisSummaryBase(test *testing.T) {
	input := []*IndividualAnalysis{
		&IndividualAnalysis{
			Score:               10,
			LinesOfCode:         10,
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
		&IndividualAnalysis{
			Score:               20,
			LinesOfCode:         40,
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
		&IndividualAnalysis{
			Score:               30,
			LinesOfCode:         20,
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
	}

	expected := &IndividualAnalysisSummary{
		AnalysisSummary: AnalysisSummary{
			Complete:       true,
			CompleteCount:  3,
			PendingCount:   0,
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

	input := []*PairwiseAnalysis{
		NewPairwiseAnalysis(NewPairwiseKey("A", "B"), sims1, nil),
		NewPairwiseAnalysis(NewPairwiseKey("C", "D"), sims2, nil),
		NewPairwiseAnalysis(NewPairwiseKey("E", "F"), sims3, nil),
	}

	expected := &PairwiseAnalysisSummary{
		AnalysisSummary: AnalysisSummary{
			Complete:       true,
			CompleteCount:  3,
			PendingCount:   0,
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
