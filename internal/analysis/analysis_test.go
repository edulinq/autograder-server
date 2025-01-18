package analysis

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestPairwiseAnalysisFake(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	oldEngines := similarityEngines
	similarityEngines = []core.SimilarityEngine{&fakeSimiliartyEngine{"fake"}}
	defer func() {
		similarityEngines = oldEngines
	}()

	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406265",
		"course101::hw0::course-student@test.edulinq.org::1697406272",
	}

	expected := []*model.PairWiseAnalysis{
		&model.PairWiseAnalysis{
			AnalysisTimestamp: timestamp.Zero(),
			SubmissionIDs: [2]string{
				"course101::hw0::course-student@test.edulinq.org::1697406256",
				"course101::hw0::course-student@test.edulinq.org::1697406265",
			},
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
			SubmissionIDs: [2]string{
				"course101::hw0::course-student@test.edulinq.org::1697406256",
				"course101::hw0::course-student@test.edulinq.org::1697406272",
			},
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
			SubmissionIDs: [2]string{
				"course101::hw0::course-student@test.edulinq.org::1697406265",
				"course101::hw0::course-student@test.edulinq.org::1697406272",
			},
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

	results, err := PairwiseAnalysis(ids)
	if err != nil {
		test.Fatalf("Failed to do pairwise analysis: '%v'.", err)
	}

	// Zero out the timestamps.
	for _, result := range results {
		result.AnalysisTimestamp = timestamp.Zero()
	}

	if !reflect.DeepEqual(expected, results) {
		test.Fatalf("Results not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(results))
	}
}

type fakeSimiliartyEngine struct {
	Name string
}

func (this *fakeSimiliartyEngine) GetName() string {
	return this.Name
}

func (this *fakeSimiliartyEngine) ComputeFileSimilarity(paths [2]string) (*model.FileSimilarity, error) {
	similarity := model.FileSimilarity{
		Filename: filepath.Base(paths[0]),
		Tool:     this.Name,
		Score:    float64(len(filepath.Base(paths[0]))) / 100.0,
	}

	return &similarity, nil
}
