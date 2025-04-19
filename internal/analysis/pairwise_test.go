package analysis

import (
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/analysis/jplag"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestPairwiseAnalysisFake(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestAssignment()

	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406265",
		"course101::hw0::course-student@test.edulinq.org::1697406272",
	}

	expected := []*model.PairwiseAnalysis{
		&model.PairwiseAnalysis{
			Options:           assignment.AssignmentAnalysisOptions,
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
			MeanSimilarities: map[string]float64{
				"submission.py": 0.13,
			},
			TotalMeanSimilarity: 0.13,
		},
		&model.PairwiseAnalysis{
			Options:           assignment.AssignmentAnalysisOptions,
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
			MeanSimilarities: map[string]float64{
				"submission.py": 0.13,
			},
			TotalMeanSimilarity: 0.13,
		},
		&model.PairwiseAnalysis{
			Options:           assignment.AssignmentAnalysisOptions,
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
			MeanSimilarities: map[string]float64{
				"submission.py": 0.13,
			},
			TotalMeanSimilarity: 0.13,
		},
	}

	testPairwise(test, ids, expected, 0)

	// Test again, which should pull from the cache.
	testPairwise(test, ids, expected, len(expected))

	query := stats.Query{
		Type: stats.MetricTypeCodeAnalysisTime,
		Where: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID: "course101",
		},
	}

	// After both runs, there should be exactly one stat record (since the second one was cached).
	results, err := db.GetMetrics(query)
	if err != nil {
		test.Fatalf("Failed to do stats query: '%v'.", err)
	}

	expectedStats := []*stats.Metric{
		&stats.Metric{
			Timestamp: timestamp.Zero(),
			Type:      stats.MetricTypeCodeAnalysisTime,
			Value:     float64(3), // 1 for each run of the fake engine.
			Attributes: map[stats.MetricAttribute]any{
				stats.MetricAttributeAnalysisType: "pairwise",
				stats.MetricAttributeCourseID:     "course101",
				stats.MetricAttributeAssignmentID: "hw0",
				stats.MetricAttributeUserEmail:    "server-admin@test.edulinq.org",
			},
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

	options := AnalysisOptions{
		ResolvedSubmissionIDs: ids,
		InitiatorEmail:        "server-admin@test.edulinq.org",
		JobOptions: jobmanager.JobOptions{
			WaitForCompletion: true,
		},
	}

	results, pendingCount, err := PairwiseAnalysis(options)
	if err != nil {
		test.Fatalf("Failed to do pairwise analysis: '%v'.", err)
	}

	if pendingCount != 0 {
		test.Fatalf("Found %d pending results, when 0 were expected.", pendingCount)
	}

	// Normalize the results.
	for _, result := range results {
		// Zero out the timestamps.
		result.AnalysisTimestamp = timestamp.Zero()

		// Nil empty skipped and unmatched files.
		for _, result := range results {
			if len(result.SkippedFiles) == 0 {
				result.SkippedFiles = nil
			}

			if len(result.UnmatchedFiles) == 0 {
				result.UnmatchedFiles = nil
			}
		}
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

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	options := AnalysisOptions{
		JobOptions: jobmanager.JobOptions{
			Context: ctx,
		},
	}

	sims, unmatches, _, _, err := computeFileSims(options, paths, nil, nil)
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

// Ensure that the default engines run.
// Full output checking will be left to the fake engine.
func TestPairwiseAnalysisDefaultEnginesBase(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	forceDefaultEnginesForTesting = true
	defer func() {
		forceDefaultEnginesForTesting = false
	}()

	defaultSimilarityEngines[1].(*jplag.JPlagEngine).MinTokens = 5
	defer func() {
		defaultSimilarityEngines[1].(*jplag.JPlagEngine).MinTokens = jplag.DEFAULT_MIN_TOKENS
	}()

	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406272",
	}

	options := AnalysisOptions{
		ResolvedSubmissionIDs: ids,
		InitiatorEmail:        "server-admin@test.edulinq.org",
		JobOptions: jobmanager.JobOptions{
			WaitForCompletion: true,
		},
	}

	results, pendingCount, err := PairwiseAnalysis(options)
	if err != nil {
		test.Fatalf("Failed to do pairwise analysis: '%v'.", err)
	}

	if pendingCount != 0 {
		test.Fatalf("Found %d pending results, when 0 were expected.", pendingCount)
	}

	if len(results) != 1 {
		test.Fatalf("Number of results not as expected. Expected: %d, Actual: %d.", 1, len(results))
	}
}

// A test for special files that seem to cause trouble with the engines.
func TestPairwiseAnalysisDefaultEnginesSpecificFiles(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	// Override a setting for JPlag for testing.
	defaultSimilarityEngines[1].(*jplag.JPlagEngine).MinTokens = 5
	defer func() {
		defaultSimilarityEngines[1].(*jplag.JPlagEngine).MinTokens = jplag.DEFAULT_MIN_TOKENS
	}()

	testPaths := []string{
		filepath.Join(util.RootDirForTesting(), "testdata", "files", "sim_engine", "config.json"),
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	for _, path := range testPaths {
		for _, engine := range defaultSimilarityEngines {
			sim, _, err := engine.ComputeFileSimilarity([2]string{path, path}, "", ctx)
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

func TestPairwiseAnalysisIncludeExclude(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
		options       *model.AssignmentAnalysisOptions
		expectedCount int
	}{
		{
			nil,
			1,
		},
		{
			&model.AssignmentAnalysisOptions{
				IncludePatterns: []string{
					`\.c$`,
				},
			},
			0,
		},
		{
			&model.AssignmentAnalysisOptions{
				ExcludePatterns: []string{
					`\.c$`,
				},
			},
			1,
		},
		{
			&model.AssignmentAnalysisOptions{
				ExcludePatterns: []string{
					`\.py$`,
				},
			},
			0,
		},
	}

	assignment := db.MustGetTestAssignment()
	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406265",
	}
	relpath := "submission.py"
	baseCount := 1

	for i, testCase := range testCases {
		db.ResetForTesting()

		if testCase.options != nil {
			err := testCase.options.Validate()
			if err != nil {
				test.Errorf("Case %d: Options is invalid: '%v'.", i, err)
				continue
			}
		}

		assignment.AssignmentAnalysisOptions = testCase.options
		db.MustSaveAssignment(assignment)

		options := AnalysisOptions{
			ResolvedSubmissionIDs: ids,
			InitiatorEmail:        "server-admin@test.edulinq.org",
			JobOptions: jobmanager.JobOptions{
				WaitForCompletion: true,
			},
		}

		results, pendingCount, err := PairwiseAnalysis(options)
		if err != nil {
			test.Errorf("Case %d: Failed to perform analysis: '%v'.", i, err)
			continue
		}

		if pendingCount != 0 {
			test.Errorf("Case %d: Found %d pending results, when 0 were expected.", i, pendingCount)
			continue
		}

		if len(results) != 1 {
			test.Errorf("Case %d: Found %d results, when 1 was expected.", i, len(results))
			continue
		}

		if testCase.expectedCount != len(results[0].Similarities) {
			test.Errorf("Case %d: Unexpected number of result similarities. Expected: %d, Actual: %d.",
				i, testCase.expectedCount, len(results[0].Similarities))
			continue
		}

		if (baseCount - testCase.expectedCount) != len(results[0].SkippedFiles) {
			test.Errorf("Case %d: Unexpected number of skipped files. Expected: %d, Actual: %d.",
				i, (baseCount - testCase.expectedCount), len(results[0].SkippedFiles))
			continue
		}

		if testCase.expectedCount == 0 {
			if relpath != results[0].SkippedFiles[0] {
				test.Errorf("Case %d: Unexpected skipped file. Expected: '%s', Actual: '%s'.",
					i, relpath, results[0].SkippedFiles[0])
			}
		}
	}
}

func TestPairwiseAnalysisCountBase(test *testing.T) {
	defer db.ResetForTesting()

	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406265",
	}

	testCases := []struct {
		options                   AnalysisOptions
		preload                   bool
		expectedCacheSetOnPreload bool
		expectedResultIsFromCache bool
		expectedResultCount       int
		expectedPendingCount      int
		expectedCacheCount        int
	}{
		// Test cases that do not wait for completion are left out because they are flaky.

		// Empty
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{},
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
			},
			preload:                   false,
			expectedCacheSetOnPreload: false,
			expectedResultIsFromCache: false,
			expectedResultCount:       0,
			expectedPendingCount:      0,
			expectedCacheCount:        0,
		},

		// Base, No Preload

		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: ids,
				JobOptions: jobmanager.JobOptions{
					OverwriteRecords:  false,
					WaitForCompletion: true,
				},
				DryRun: false,
			},
			preload:                   false,
			expectedCacheSetOnPreload: false,
			expectedResultIsFromCache: false,
			expectedResultCount:       1,
			expectedPendingCount:      0,
			expectedCacheCount:        1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: ids,
				JobOptions: jobmanager.JobOptions{
					OverwriteRecords:  false,
					WaitForCompletion: true,
				},
				DryRun: true,
			},
			preload:                   false,
			expectedCacheSetOnPreload: false,
			expectedResultIsFromCache: false,
			expectedResultCount:       1,
			expectedPendingCount:      0,
			expectedCacheCount:        0,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: ids,
				JobOptions: jobmanager.JobOptions{
					OverwriteRecords:  true,
					WaitForCompletion: true,
				},
				DryRun: false,
			},
			preload:                   false,
			expectedCacheSetOnPreload: false,
			expectedResultIsFromCache: false,
			expectedResultCount:       1,
			expectedPendingCount:      0,
			expectedCacheCount:        1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: ids,
				JobOptions: jobmanager.JobOptions{
					OverwriteRecords:  true,
					WaitForCompletion: true,
				},
				DryRun: true,
			},
			preload:                   false,
			expectedCacheSetOnPreload: false,
			expectedResultIsFromCache: false,
			expectedResultCount:       1,
			expectedPendingCount:      0,
			expectedCacheCount:        0,
		},

		// Base, Preload

		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: ids,
				JobOptions: jobmanager.JobOptions{
					OverwriteRecords:  false,
					WaitForCompletion: true,
				},
				DryRun: false,
			},
			preload:                   true,
			expectedCacheSetOnPreload: true,
			expectedResultIsFromCache: true,
			expectedResultCount:       1,
			expectedPendingCount:      0,
			expectedCacheCount:        1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: ids,
				JobOptions: jobmanager.JobOptions{
					OverwriteRecords:  false,
					WaitForCompletion: true,
				},
				DryRun: true,
			},
			preload:                   true,
			expectedCacheSetOnPreload: true,
			expectedResultIsFromCache: true,
			expectedResultCount:       1,
			expectedPendingCount:      0,
			expectedCacheCount:        1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: ids,
				JobOptions: jobmanager.JobOptions{
					OverwriteRecords:  true,
					WaitForCompletion: true,
				},
				DryRun: false,
			},
			preload:                   true,
			expectedCacheSetOnPreload: false,
			expectedResultIsFromCache: false,
			expectedResultCount:       1,
			expectedPendingCount:      0,
			expectedCacheCount:        1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: ids,
				JobOptions: jobmanager.JobOptions{
					OverwriteRecords:  true,
					WaitForCompletion: true,
				},
				DryRun: true,
			},
			preload:                   true,
			expectedCacheSetOnPreload: true,
			expectedResultIsFromCache: false,
			expectedResultCount:       1,
			expectedPendingCount:      0,
			expectedCacheCount:        1,
		},
	}

	// This test will need strong context control since we are not waiting for all the results.
	var ctx context.Context = nil
	var contextCancelFunc context.CancelFunc = nil

	defer func() {
		if contextCancelFunc != nil {
			contextCancelFunc()
		}
	}()

	for i, testCase := range testCases {
		// Cancel any old runs.
		if contextCancelFunc != nil {
			contextCancelFunc()
		}

		db.ResetForTesting()

		if testCase.preload {
			preloadOptions := AnalysisOptions{
				ResolvedSubmissionIDs: testCase.options.ResolvedSubmissionIDs,
				InitiatorEmail:        "server-admin@test.edulinq.org",
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
			}

			_, _, err := PairwiseAnalysis(preloadOptions)
			if err != nil {
				test.Errorf("Case %d: Failed to preload analysis: '%v'.", i, err)
				continue
			}
		}

		// Mark now and sleep for a very small amount of time.
		time.Sleep(time.Duration(5) * time.Millisecond)
		startTime := timestamp.Now()
		time.Sleep(time.Duration(5) * time.Millisecond)

		// Create a new cancellable context for this run.
		ctx, contextCancelFunc = context.WithCancel(context.Background())

		testCase.options.InitiatorEmail = "server-admin@test.edulinq.org"
		testCase.options.JobOptions.Context = ctx
		testCase.options.JobOptions.RetainOriginalContext = false

		results, pendingCount, err := PairwiseAnalysis(testCase.options)
		if err != nil {
			test.Errorf("Case %d: Failed to do analysis: '%v'.", i, err)
			continue
		}

		if testCase.expectedResultCount != len(results) {
			test.Errorf("Case %d: Unexpected number of results. Expected: %d, Actual: %d.",
				i, testCase.expectedResultCount, len(results))
			continue
		}

		if testCase.expectedPendingCount != pendingCount {
			test.Errorf("Case %d: Unexpected number of pending results. Expected: %d, Actual: %d.",
				i, testCase.expectedPendingCount, pendingCount)
			continue
		}

		// Check if the result was from the cache using the start time.
		if len(results) > 0 {
			resultTime := results[0].AnalysisTimestamp
			resultIsFromCache := (resultTime <= startTime)

			if testCase.expectedResultIsFromCache != resultIsFromCache {
				test.Errorf("Case %d: Unexpected result being from cache. Expected: %v, Actual: %v.",
					i, testCase.expectedResultIsFromCache, resultIsFromCache)
				continue
			}
		}

		dbResults, err := db.GetPairwiseAnalysis(createPairwiseKeys(testCase.options.ResolvedSubmissionIDs))
		if err != nil {
			test.Errorf("Case %d: Failed to do get db results: '%v'.", i, err)
			continue
		}

		if testCase.expectedCacheCount != len(dbResults) {
			test.Errorf("Case %d: Unexpected number of db results. Expected: %d, Actual: %d.",
				i, testCase.expectedCacheCount, len(dbResults))
			continue
		}

		if len(dbResults) == 0 {
			continue
		}

		// Check if the cache was set on preload by comparing the analysis time.
		key := model.NewPairwiseKey(testCase.options.ResolvedSubmissionIDs[0], testCase.options.ResolvedSubmissionIDs[1])
		cacheTime := dbResults[key].AnalysisTimestamp
		if testCase.expectedCacheSetOnPreload {
			if cacheTime > startTime {
				test.Errorf("Case %d: Cache entry was set after preload.", i)
				continue
			}
		} else {
			if cacheTime < startTime {
				test.Errorf("Case %d: Cache entry was set during preload.", i)
				continue
			}
		}
	}
}

func TestPairwiseAnalysisFailureBase(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testFailPairwiseAnalysis = true
	defer func() {
		testFailPairwiseAnalysis = false
	}()

	expectedMessageSubstring := "Test failure."

	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406256",
		"course101::hw0::course-student@test.edulinq.org::1697406272",
	}

	options := AnalysisOptions{
		ResolvedSubmissionIDs: ids,
		InitiatorEmail:        "server-admin@test.edulinq.org",
		JobOptions: jobmanager.JobOptions{
			WaitForCompletion: true,
		},
	}

	results, pendingCount, err := PairwiseAnalysis(options)
	if err != nil {
		test.Fatalf("Failed to do pairwise analysis: '%v'.", err)
	}

	if pendingCount != 0 {
		test.Fatalf("Found %d pending results, when 0 were expected.", pendingCount)
	}

	if len(results) != 1 {
		test.Fatalf("Number of results not as expected. Expected: %d, Actual: %d.", 1, len(results))
	}

	if !results[0].Failure {
		test.Fatalf("Result is not a failure, when it should be.")
	}

	if !strings.Contains(results[0].FailureMessage, expectedMessageSubstring) {
		test.Fatalf("Failure message does not contain expected substring. Expected Substring: '%s', Actual: '%s'.", expectedMessageSubstring, results[0].FailureMessage)
	}
}
