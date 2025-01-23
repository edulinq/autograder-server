package analysis

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/analysis/dolos"
	"github.com/edulinq/autograder/internal/analysis/jplag"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestPairwiseAnalysisDefaultEnginesBase(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	db.ResetForTesting()
	defer db.ResetForTesting()

	// Force the use of the default engines.
	forceDefaultEnginesForTesting = true
	defer func() {
		forceDefaultEnginesForTesting = false
	}()

	// Override a setting for JPlag for testing.
	defaultSimilarityEngines[1].(*jplag.JPlagEngine).MinTokens = 5

	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406265",
		"course101::hw0::course-student@test.edulinq.org::1697406272",
	}

	expected := []*model.PairwiseAnalysis{
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
						Tool:     dolos.NAME,
						Version:  dolos.VERSION,
						Score:    0,
					},
					&model.FileSimilarity{
						Filename: "submission.py",
						Tool:     jplag.NAME,
						Version:  jplag.VERSION,
						Score:    0,
					},
				},
			},
			UnmatchedFiles: [][2]string{},
			MeanSimilarities: map[string]float64{
				"submission.py": 0,
			},
			TotalMeanSimilarity: 0,
		},
	}

	results, pendingCount, err := PairwiseAnalysis(ids, true, "course-admin@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to do pairwise analysis: '%v'.", err)
	}

	if pendingCount != 0 {
		test.Fatalf("Found %d pending results, when 0 were expected.", pendingCount)
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

		meanSim := result.MeanSimilarities["submission.py"]
		result.MeanSimilarities["submission.py"] = 0
		if meanSim <= 0 {
			test.Fatalf("Mean similairty is not above zero, found %f.", meanSim)
		}

		totalMeanSim := result.TotalMeanSimilarity
		result.TotalMeanSimilarity = 0
		if totalMeanSim <= 0 {
			test.Fatalf("Total mean similairty is not above zero, found %f.", totalMeanSim)
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

	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406265",
		"course101::hw0::course-student@test.edulinq.org::1697406272",
	}

	expected := []*model.PairwiseAnalysis{
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
			MeanSimilarities: map[string]float64{
				"submission.py": 0.13,
			},
			TotalMeanSimilarity: 0.13,
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
			MeanSimilarities: map[string]float64{
				"submission.py": 0.13,
			},
			TotalMeanSimilarity: 0.13,
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
			MeanSimilarities: map[string]float64{
				"submission.py": 0.13,
			},
			TotalMeanSimilarity: 0.13,
		},
	}

	testPairwise(test, ids, expected, 0)

	// Test again, which should pull from the cache.
	testPairwise(test, ids, expected, len(expected))

	// After both runs, there should be exactly one stat record (since the second one was cached).
	results, err := db.GetCourseMetrics(stats.CourseMetricQuery{CourseID: "course101"})
	if err != nil {
		test.Fatalf("Failed to do stats query: '%v'.", err)
	}

	expectedStats := []*stats.CourseMetric{
		&stats.CourseMetric{
			BaseMetric: stats.BaseMetric{
				Timestamp: timestamp.Zero(),
				Attributes: map[string]any{
					"anslysis-type": "pairwise",
				},
			},
			Type:         stats.CourseMetricTypeCodeAnalysisTime,
			CourseID:     "course101",
			AssignmentID: "hw0",
			UserEmail:    "server-admin@test.edulinq.org",
			Value:        3, // 1 for each run of the fake engine.
		},
	}

	// Zero out the query results.
	for _, result := range results {
		result.Timestamp = timestamp.Zero()
	}

	if !reflect.DeepEqual(expectedStats, results) {
		test.Fatalf("Stat results not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expectedStats), util.MustToJSONIndent(results))
	}
}

func testPairwise(test *testing.T, ids []string, expected []*model.PairwiseAnalysis, expectedInitialCacheCount int) {
	// Check for records in the DB.
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

	results, pendingCount, err := PairwiseAnalysis(ids, true, "server-admin@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to do pairwise analysis: '%v'.", err)
	}

	if pendingCount != 0 {
		test.Fatalf("Found %d pending results, when 0 were expected.", pendingCount)
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

func TestPairwiseWithPythonNotebook(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	tempDir := util.MustMkDirTemp("test-analysis-pairwise-")
	defer util.RemoveDirent(tempDir)

	err := util.CopyDir(filepath.Join(util.RootDirForTesting(), "testdata", "files", "python_notebook"), tempDir, true)
	if err != nil {
		test.Fatalf("Failed to prep temp dir: '%v'.", err)
	}

	paths := [2]string{
		filepath.Join(tempDir, "ipynb"),
		filepath.Join(tempDir, "py"),
	}

	sims, unmatches, _, err := computeFileSims(paths, "test")
	if err != nil {
		test.Fatalf("Failed to compute file similarity: '%v'.", err)
	}

	if len(unmatches) != 0 {
		test.Fatalf("Unexpected number of unmatches. Expected: 0, Actual: %d, Unmatches: '%s'.", len(unmatches), util.MustToJSONIndent(unmatches))
	}

	expected := map[string][]*model.FileSimilarity{
		"submission.py": []*model.FileSimilarity{
			&model.FileSimilarity{
				Filename:         "submission.py",
				OriginalFilename: "submission.ipynb",
				Tool:             "fake",
				Version:          "0.0.1",
				Score:            0.13,
			},
		},
	}

	if !reflect.DeepEqual(expected, sims) {
		test.Fatalf("Results not as expected. Expected: '%s', Actual: '%s'.", util.MustToJSONIndent(expected), util.MustToJSONIndent(sims))
	}
}

// A test for special files that seem to cause trouble with the engines.
func TestPairwiseAnalysisDefaultEnginesSpecificFiles(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	// Override a setting for JPlag for testing.
	defaultSimilarityEngines[1].(*jplag.JPlagEngine).MinTokens = 5

	testPaths := []string{
		filepath.Join(util.RootDirForTesting(), "testdata", "files", "sim_engine", "config.json"),
	}

	for _, path := range testPaths {
		for _, engine := range defaultSimilarityEngines {
			sim, _, err := engine.ComputeFileSimilarity([2]string{path, path}, "test")
			if err != nil {
				test.Errorf("Engine '%s' failed to compute similarity on '%s': '%v'.",
					engine.GetName(), path, err)
				continue
			}

			expected := 1.0
			if !util.IsClose(expected, sim.Score) {
				test.Errorf("Engine '%s' got an unexpected score on self-similarity with '%s'. Expected: %f, Actual: %f.",
					engine.GetName(), path, expected, sim.Score)
				continue
			}
		}
	}
}
