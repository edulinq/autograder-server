package analysis

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestPairwiseAnalysisDefaultEngines(test *testing.T) {
	// The Dolos container has some strange permission issues when run on Github Actions.
	if config.GITHUB_CI.Get() {
		test.Skip("Skipping on Github Actions.")
	}

	docker.EnsureOrSkipForTest(test)

	db.ResetForTesting()
	defer db.ResetForTesting()

	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406265",
		"course101::hw0::course-student@test.edulinq.org::1697406272",
	}

	expected := []*model.PairWiseAnalysis{
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
						Tool:     "dolos",
						Version:  "2.9.0",
						Score:    0,
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

	for _, result := range results {
		// Zero out the timestamps.
		result.AnalysisTimestamp = timestamp.Zero()

		// We only care that the scores are above zero.
		for i, similarity := range result.Similarities["submission.py"] {
			if similarity.Score <= 0 {
				test.Fatalf("Similairty from '%s' (index %d) is not above zero, found %f.", similarity.Tool, i, similarity.Score)
			}

			// Zero out score after check.
			similarity.Score = 0
		}
	}

	if !reflect.DeepEqual(expected, results) {
		test.Fatalf("Results not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(results))
	}
}

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

	testPairwise(test, ids, expected, 0)

	// Test again, which should pull from the cache.
	testPairwise(test, ids, expected, len(expected))
}

func testPairwise(test *testing.T, ids []string, expected []*model.PairWiseAnalysis, expectedInitialCacheCount int) {
	// Ensure that there are no records for these in the DB.
	queryKeys := make([]model.PairwiseKey, 0, len(expected))
	for _, analysis := range expected {
		queryKeys = append(queryKeys, analysis.SubmissionIDs)
	}

	queryResult, err := db.GetPairwiseAnalysis(queryKeys)
	if err != nil {
		test.Fatalf("Failed to do initial query for cached anslysis: '%v'.", err)
	}

	if len(queryResult) != expectedInitialCacheCount {
		test.Fatalf("Number of (pre) cached anslysis results not as expected. Expected: %d, Actual: %d.", expectedInitialCacheCount, len(queryResult))
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

	queryResult, err = db.GetPairwiseAnalysis(queryKeys)
	if err != nil {
		test.Fatalf("Failed to do query for cached anslysis: '%v'.", err)
	}

	if len(queryResult) != len(expected) {
		test.Fatalf("Number of (post) cached anslysis results not as expected. Expected: %d, Actual: %d.", len(expected), len(queryResult))
	}
}

type fakeSimiliartyEngine struct {
	Name string
}

func (this *fakeSimiliartyEngine) GetName() string {
	return this.Name
}

func (this *fakeSimiliartyEngine) ComputeFileSimilarity(paths [2]string, baseLockKey string) (*model.FileSimilarity, error) {
	similarity := model.FileSimilarity{
		Filename: filepath.Base(paths[0]),
		Tool:     this.Name,
		Score:    float64(len(filepath.Base(paths[0]))) / 100.0,
	}

	return &similarity, nil
}
