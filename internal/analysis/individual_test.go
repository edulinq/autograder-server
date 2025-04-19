package analysis

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestIndividualAnalysisBase(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestAssignment()

	ids := []string{
		"course101::hw0::course-student@test.edulinq.org::1697406265",
	}

	expected := []*model.IndividualAnalysis{
		&model.IndividualAnalysis{
			AnalysisTimestamp: timestamp.Zero(),
			Options:           assignment.AssignmentAnalysisOptions,

			FullID:       ids[0],
			ShortID:      "1697406265",
			CourseID:     "course101",
			AssignmentID: "hw0",
			UserEmail:    "course-student@test.edulinq.org",

			SubmissionStartTime: timestamp.FromMSecs(1697406266000),
			Score:               1,

			Files: []model.AnalysisFileInfo{
				model.AnalysisFileInfo{
					Filename:    "submission.py",
					LinesOfCode: 4,
				},
			},
			LinesOfCode: 4,

			SubmissionTimeDelta: 10000,
			LinesOfCodeDelta:    0,
			ScoreDelta:          1,

			LinesOfCodeVelocity: 0,
			ScoreVelocity:       360,
		},
	}

	testIndividual(test, ids, expected, 0)

	// Test again, which should pull from the cache.
	testIndividual(test, ids, expected, len(expected))

	query := stats.Query{
		Type: stats.MetricTypeCodeAnalysisTime,
		Where: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID: db.TEST_COURSE_ID,
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
			Value:     0,
			Attributes: map[stats.MetricAttribute]any{
				stats.MetricAttributeAnalysisType: "individual",
				stats.MetricAttributeCourseID:     "course101",
				stats.MetricAttributeAssignmentID: "hw0",
				stats.MetricAttributeUserEmail:    "server-admin@test.edulinq.org",
			},
		},
	}

	// Zero out the query results.
	for _, result := range results {
		result.Timestamp = timestamp.Zero()

		if result.Attributes == nil {
			result.Attributes = make(map[stats.MetricAttribute]any)
		}

		result.Value = 0
	}

	if !reflect.DeepEqual(expectedStats, results) {
		test.Fatalf("Stat results not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expectedStats), util.MustToJSONIndent(results))
	}
}

func testIndividual(test *testing.T, ids []string, expected []*model.IndividualAnalysis, expectedInitialCacheCount int) {
	queryResult, err := db.GetIndividualAnalysis(ids)
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

	results, pendingCount, err := IndividualAnalysis(options)
	if err != nil {
		test.Fatalf("Failed to do individual analysis: '%v'.", err)
	}

	if pendingCount != 0 {
		test.Fatalf("Found %d pending results, when 0 were expected.", pendingCount)
	}

	// Normalize the results.
	for _, result := range results {
		// Zero out the timestamps.
		result.AnalysisTimestamp = timestamp.Zero()

		// Nil empty skipped files.
		for _, result := range results {
			if len(result.SkippedFiles) == 0 {
				result.SkippedFiles = nil
			}
		}
	}

	if !reflect.DeepEqual(expected, results) {
		test.Fatalf("Results not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(results))
	}

	queryResult, err = db.GetIndividualAnalysis(ids)
	if err != nil {
		test.Fatalf("Failed to do query for cached anslysis: '%v'.", err)
	}

	if len(queryResult) != len(expected) {
		test.Fatalf("Number of (post) cached anslysis results not as expected. Expected: %d, Actual: %d.", len(expected), len(queryResult))
	}
}

func TestIndividualAnalysisIncludeExclude(test *testing.T) {
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
	submissionIDs := []string{"course101::hw0::course-student@test.edulinq.org::1697406265"}
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
			ResolvedSubmissionIDs: submissionIDs,
			InitiatorEmail:        "server-admin@test.edulinq.org",
			JobOptions: jobmanager.JobOptions{
				WaitForCompletion: true,
			},
		}

		results, pendingCount, err := IndividualAnalysis(options)
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

		if testCase.expectedCount != len(results[0].Files) {
			test.Errorf("Case %d: Unexpected number of result files. Expected: %d, Actual: %d.",
				i, testCase.expectedCount, len(results[0].Files))
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
				continue
			}
		}
	}
}

func TestIndividualAnalysisCountBase(test *testing.T) {
	defer db.ResetForTesting()

	submissionID := "course101::hw0::course-student@test.edulinq.org::1697406256"

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
				ResolvedSubmissionIDs: []string{
					submissionID,
				},
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
				ResolvedSubmissionIDs: []string{
					submissionID,
				},
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
				ResolvedSubmissionIDs: []string{
					submissionID,
				},
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
				ResolvedSubmissionIDs: []string{
					submissionID,
				},
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
				ResolvedSubmissionIDs: []string{
					submissionID,
				},
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
				ResolvedSubmissionIDs: []string{
					submissionID,
				},
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
				ResolvedSubmissionIDs: []string{
					submissionID,
				},
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
				ResolvedSubmissionIDs: []string{
					submissionID,
				},
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
		db.ResetForTesting()

		// Cancel any old runs.
		if contextCancelFunc != nil {
			contextCancelFunc()
		}

		if testCase.preload {
			preloadOptions := AnalysisOptions{
				ResolvedSubmissionIDs: testCase.options.ResolvedSubmissionIDs,
				InitiatorEmail:        "server-admin@test.edulinq.org",
				JobOptions: jobmanager.JobOptions{
					WaitForCompletion: true,
				},
			}

			_, _, err := IndividualAnalysis(preloadOptions)
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

		results, pendingCount, err := IndividualAnalysis(testCase.options)
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

		dbResults, err := db.GetIndividualAnalysis(testCase.options.ResolvedSubmissionIDs)
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
		cacheTime := dbResults[testCase.options.ResolvedSubmissionIDs[0]].AnalysisTimestamp
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

func TestIndividualAnalysisFailureBase(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testFailIndividualAnalysis = true
	defer func() {
		testFailIndividualAnalysis = false
	}()

	expectedMessageSubstring := "Test failure."

	options := AnalysisOptions{
		ResolvedSubmissionIDs: []string{"course101::hw0::course-student@test.edulinq.org::1697406265"},
		InitiatorEmail:        "server-admin@test.edulinq.org",
		JobOptions: jobmanager.JobOptions{
			WaitForCompletion: true,
		},
	}

	results, pendingCount, err := IndividualAnalysis(options)
	if err != nil {
		test.Fatalf("Failed to perform analysis: '%v'.", err)
	}

	if pendingCount != 0 {
		test.Fatalf("Found %d pending results, when 0 were expected.", pendingCount)
	}

	if len(results) != 1 {
		test.Fatalf("Found %d results, when 1 was expected.", len(results))
	}

	if !results[0].Failure {
		test.Fatalf("Result is not a failure, when it should be.")
	}

	if !strings.Contains(results[0].FailureMessage, expectedMessageSubstring) {
		test.Fatalf("Failure message does not contain expected substring. Expected Substring: '%s', Actual: '%s'.", expectedMessageSubstring, results[0].FailureMessage)
	}
}
