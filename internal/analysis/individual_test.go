package analysis

import (
	"reflect"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/db"
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
			Options:           assignment.AnalysisOptions,

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
			SkippedFiles: []string{},
			LinesOfCode:  4,

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
					stats.ATTRIBUTE_KEY_ANALYSIS: "individual",
				},
			},
			Type:         stats.CourseMetricTypeCodeAnalysisTime,
			CourseID:     "course101",
			AssignmentID: "hw0",
			UserEmail:    "server-admin@test.edulinq.org",
			Value:        0,
		},
	}

	// Zero out the query results.
	for _, result := range results {
		result.Timestamp = timestamp.Zero()
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
		WaitForCompletion:     true,
	}

	results, pendingCount, err := IndividualAnalysis(options, "server-admin@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to do individual analysis: '%v'.", err)
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
		options       *model.AnalysisOptions
		expectedCount int
	}{
		{
			nil,
			1,
		},
		{
			&model.AnalysisOptions{
				IncludePatterns: []string{
					`\.c$`,
				},
			},
			0,
		},
		{
			&model.AnalysisOptions{
				ExcludePatterns: []string{
					`\.c$`,
				},
			},
			1,
		},
		{
			&model.AnalysisOptions{
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

		assignment.AnalysisOptions = testCase.options
		db.MustSaveAssignment(assignment)

		options := AnalysisOptions{
			ResolvedSubmissionIDs: submissionIDs,
			WaitForCompletion:     true,
		}

		results, pendingCount, err := IndividualAnalysis(options, "server-admin@test.edulinq.org")
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
			}
		}
	}
}

func TestIndividualAnalysisCountBase(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		options              AnalysisOptions
		preload              bool
		wait                 bool
		expectedResultCount  int
		expectedPendingCount int
		expectedCacheCount   int
	}{
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{
					"course101::hw0::course-student@test.edulinq.org::1697406256",
				},
				WaitForCompletion: true,
			},
			preload:              false,
			expectedResultCount:  1,
			expectedPendingCount: 0,
			expectedCacheCount:   1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{},
				WaitForCompletion:     true,
			},
			preload:              false,
			expectedResultCount:  0,
			expectedPendingCount: 0,
			expectedCacheCount:   0,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{
					"course101::hw0::course-student@test.edulinq.org::1697406256",
				},
				WaitForCompletion: true,
				DryRun:            true,
			},
			preload:              false,
			expectedResultCount:  1,
			expectedPendingCount: 0,
			expectedCacheCount:   0,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{
					"course101::hw0::course-student@test.edulinq.org::1697406256",
				},
				WaitForCompletion: true,
				OverwriteCache:    true,
			},
			preload:              false,
			expectedResultCount:  1,
			expectedPendingCount: 0,
			expectedCacheCount:   1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{
					"course101::hw0::course-student@test.edulinq.org::1697406256",
				},
				WaitForCompletion: true,
				OverwriteCache:    true,
				DryRun:            true,
			},
			preload:              false,
			expectedResultCount:  1,
			expectedPendingCount: 0,
			expectedCacheCount:   0,
		},

		// Preload

		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{
					"course101::hw0::course-student@test.edulinq.org::1697406256",
				},
				WaitForCompletion: false,
			},
			preload:              true,
			expectedResultCount:  1,
			expectedPendingCount: 0,
			expectedCacheCount:   1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{},
				WaitForCompletion:     false,
			},
			preload:              true,
			expectedResultCount:  0,
			expectedPendingCount: 0,
			expectedCacheCount:   0,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{
					"course101::hw0::course-student@test.edulinq.org::1697406256",
				},
				WaitForCompletion: false,
				DryRun:            true,
			},
			preload:              true,
			expectedResultCount:  1,
			expectedPendingCount: 0,
			expectedCacheCount:   1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{
					"course101::hw0::course-student@test.edulinq.org::1697406256",
				},
				WaitForCompletion: false,
				OverwriteCache:    true,
			},
			preload:              true,
			wait:                 true,
			expectedResultCount:  0,
			expectedPendingCount: 1,
			expectedCacheCount:   1,
		},
		{
			options: AnalysisOptions{
				ResolvedSubmissionIDs: []string{
					"course101::hw0::course-student@test.edulinq.org::1697406256",
				},
				WaitForCompletion: false,
				OverwriteCache:    true,
				DryRun:            true,
			},
			preload:              true,
			wait:                 true,
			expectedResultCount:  0,
			expectedPendingCount: 1,
			expectedCacheCount:   1,
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		if testCase.preload {
			preloadOptions := AnalysisOptions{
				ResolvedSubmissionIDs: testCase.options.ResolvedSubmissionIDs,
				WaitForCompletion:     true,
			}

			_, _, err := IndividualAnalysis(preloadOptions, "server-admin@test.edulinq.org")
			if err != nil {
				test.Errorf("Case %d: Failed to preload analysis: '%v'.", i, err)
				continue
			}
		}

		results, pendingCount, err := IndividualAnalysis(testCase.options, "server-admin@test.edulinq.org")
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

		// Wait long enough for the analysis to finish.
		if testCase.wait {
			time.Sleep(time.Duration(100) * time.Millisecond)
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
	}
}
