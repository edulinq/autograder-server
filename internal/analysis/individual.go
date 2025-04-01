package analysis

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

var testFailIndividualAnalysis bool = false

func IndividualAnalysis(options AnalysisOptions) ([]*model.IndividualAnalysis, int, error) {
	if options.Context == nil {
		options.Context = context.Background()
	}

	completeAnalysis, remainingIDs, err := getCachedIndividualResults(options)
	if err != nil {
		return nil, 0, err
	}

	// TEST
	fmt.Printf("TEST - %s - Start individual: '%s'.\n", options.InitiatorEmail, util.MustToJSON(options))

	if options.WaitForCompletion {
		results, err := runIndividualAnalysis(options, remainingIDs)
		if err != nil {
			return nil, 0, err
		}

		completeAnalysis = append(completeAnalysis, results...)
		remainingIDs = nil
	} else {
		if !options.RetainOriginalContext {
			options.Context = context.Background()
		}

		go func() {
			_, err := runIndividualAnalysis(options, remainingIDs)
			if err != nil {
				log.Error("Failure during asynchronous individual analysis.", err)
			}
		}()
	}

	return completeAnalysis, len(remainingIDs), nil
}

func getCachedIndividualResults(options AnalysisOptions) ([]*model.IndividualAnalysis, []string, error) {
	// Sort the ids so the result will be consistently ordered.
	fullSubmissionIDs := slices.Clone(options.ResolvedSubmissionIDs)
	slices.Sort(fullSubmissionIDs)

	// If we are overwriting the cache, don't query the DB for any of the cached results.
	if options.OverwriteCache {
		return make([]*model.IndividualAnalysis, 0), fullSubmissionIDs, nil
	}

	return getCachedIndividualResultsInternal(fullSubmissionIDs)
}

func getCachedIndividualResultsInternal(fullSubmissionIDs []string) ([]*model.IndividualAnalysis, []string, error) {
	// Get any already done analysis results from the DB.
	dbResults, err := db.GetIndividualAnalysis(fullSubmissionIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get cached individual analysis from DB: '%w'.", err)
	}

	// Split up the IDs into complete and remaining.
	completeAnalysis := make([]*model.IndividualAnalysis, 0, len(dbResults))
	remainingIDs := make([]string, 0, len(fullSubmissionIDs)-len(dbResults))

	for _, id := range fullSubmissionIDs {
		result, ok := dbResults[id]
		if ok {
			completeAnalysis = append(completeAnalysis, result)
		} else {
			remainingIDs = append(remainingIDs, id)
		}
	}

	return completeAnalysis, remainingIDs, nil
}

// Lock based on course and then run the analysis in a parallel pool.
func runIndividualAnalysis(options AnalysisOptions, fullSubmissionIDs []string) ([]*model.IndividualAnalysis, error) {
	if len(fullSubmissionIDs) == 0 {
		// TEST
		fmt.Printf("TEST - %s - runIndividualAnalysis got everything from the cache.\n", options.InitiatorEmail)

		return nil, nil
	}

	// Lock based on the first seen course.
	// This is to prevent multiple requests using up all the cores.
	lockCourseID, _, _, _, err := common.SplitFullSubmissionID(fullSubmissionIDs[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to get locking course: '%w'.", err)
	}

	// TEST
	fmt.Printf("TEST - %s - runIndividualAnalysis initial keys: '%v'.\n", options.InitiatorEmail, fullSubmissionIDs)
	fmt.Printf("TEST - %s - runIndividualAnalysis lock: '%s'.\n", options.InitiatorEmail, lockCourseID)

	lockKey := fmt.Sprintf("analysis-individual-course-%s", lockCourseID)
	noLockWait := lockmanager.Lock(lockKey)
	defer lockmanager.Unlock(lockKey)

	// TEST
	fmt.Printf("TEST - %s - runIndividualAnalysis unlock: '%s'.\n", options.InitiatorEmail, lockCourseID)

	// The context has been canceled while waiting for a lock, abandon this analysis.
	if options.Context.Err() != nil {
		return nil, nil
	}

	results := make([]*model.IndividualAnalysis, 0, len(fullSubmissionIDs))

	// If we had to wait for the lock, then check again for cached results.
	// If there are multiple requests queued up,
	// it will be faster to do a bulk check for cached results instead of checking each one individually.
	if !noLockWait {
		// TEST
		fmt.Printf("TEST - %s - runIndividualAnalysis fetching keys.\n", options.InitiatorEmail)

		var partialResults []*model.IndividualAnalysis = nil
		partialResults, fullSubmissionIDs, err = getCachedIndividualResultsInternal(fullSubmissionIDs)
		if err != nil {
			return nil, fmt.Errorf("Failed to re-check result cache before run: '%w'.", err)
		}

		// TEST
		fmt.Printf("TEST - %s - runIndividualAnalysis (second) cache results: '%s', '%s'.\n", options.InitiatorEmail, util.MustToJSON(partialResults), util.MustToJSON(fullSubmissionIDs))

		// Collect the partial results from the cache.
		results = append(results, partialResults...)
	}

	// All results were fetched from the cache.
	if len(fullSubmissionIDs) == 0 {
		// TEST
		fmt.Printf("TEST - %s - runIndividualAnalysis all results fetched from cache (second-level): '%v'.\n", options.InitiatorEmail, util.MustToJSON(results))

		return results, nil
	}

	// TEST
	fmt.Printf("TEST - %s - runIndividualAnalysis ids: '%v'.\n", options.InitiatorEmail, fullSubmissionIDs)

	// If we are overwriting the cache, then remove all the old entries.
	if options.OverwriteCache && !options.DryRun {
		err := db.RemoveIndividualAnalysis(fullSubmissionIDs)
		if err != nil {
			fmt.Errorf("Failed to remove old individual analysis cache entries: '%w'.", err)
		}
	}

	poolSize := config.ANALYSIS_INDIVIDUAL_COURSE_POOL_SIZE.Get()
	type PoolResult struct {
		Result  *model.IndividualAnalysis
		RunTime int64
		Error   error
	}

	poolResults, _, err := util.RunParallelPoolMap(poolSize, fullSubmissionIDs, options.Context, func(fullSubmissionID string) (PoolResult, error) {
		result, runTime, err := runSingleIndividualAnalysis(options, fullSubmissionID)
		if err != nil {
			err = fmt.Errorf("Failed to perform individual analysis on submission %s: '%w'.", fullSubmissionID, err)
		}

		return PoolResult{result, runTime, err}, nil
	})

	// If the analysis was canceled, exit right away.
	if options.Context.Err() != nil {
		return nil, nil
	}

	// TEST
	fmt.Printf("TEST - %s - runIndividualAnalysis pool results: '%v'.\n", options.InitiatorEmail, util.MustToJSON(poolResults))

	var errs error = nil
	totalRunTime := int64(0)

	for _, poolResult := range poolResults {
		if poolResult.Error != nil {
			errs = errors.Join(errs, poolResult.Error)
		} else {
			results = append(results, poolResult.Result)
			totalRunTime += poolResult.RunTime
		}
	}

	collectIndividualStats(fullSubmissionIDs, totalRunTime, options.InitiatorEmail)

	return results, errs
}

func runSingleIndividualAnalysis(options AnalysisOptions, fullSubmissionID string) (*model.IndividualAnalysis, int64, error) {
	// TEST
	fmt.Printf("TEST - %s - runSingleIndividualAnalysis lock: '%s'.\n", options.InitiatorEmail, fullSubmissionID)

	// Lock this id so we don't try to do the analysis multiple times.
	lockKey := fmt.Sprintf("analysis-individual-%s", fullSubmissionID)
	lockmanager.Lock(lockKey)
	defer lockmanager.Unlock(lockKey)

	// TEST
	fmt.Printf("TEST - %s - runSingleIndividualAnalysis unlock: '%s'.\n", options.InitiatorEmail, fullSubmissionID)

	// The context has been canceled while waiting for a lock, abandon this analysis.
	if options.Context.Err() != nil {
		return nil, 0, nil
	}

	startTime := timestamp.Now()

	// Check the DB for a complete analysis.
	if !options.OverwriteCache {
		result, err := db.GetSingleIndividualAnalysis(fullSubmissionID)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to check DB for cached individual analysis for '%s': '%w'.", fullSubmissionID, err)
		}

		if result != nil {
			// TEST
			fmt.Printf("TEST - %s - runSingleIndividualAnalysis cache hit!: '%v'.\n", options.InitiatorEmail, util.MustToJSON(result))

			return result, 0, nil
		}
	}

	// Nothing cached, compute the analysis.
	result, err := computeSingleIndividualAnalysis(options, fullSubmissionID, true)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to compute individual analysis for '%s': '%w'.", fullSubmissionID, err)
	}

	// TEST
	fmt.Printf("TEST - %s - runSingleIndividualAnalysis result: '%v'.\n", options.InitiatorEmail, util.MustToJSON(result))

	// Store the result.
	if !options.DryRun && (options.Context.Err() == nil) {
		err = db.StoreIndividualAnalysis([]*model.IndividualAnalysis{result})
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to store individual analysis for '%s' in DB: '%w'.", fullSubmissionID, err)
		}
	}

	runTime := int64(timestamp.Now() - startTime)

	return result, runTime, nil
}

func computeSingleIndividualAnalysis(options AnalysisOptions, fullSubmissionID string, computeDeltas bool) (*model.IndividualAnalysis, error) {
	tempDir, err := util.MkDirTemp("individual-analysis-")
	if err != nil {
		return nil, fmt.Errorf("Failed to make temp dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	submissionDir := filepath.Join(tempDir, "current")
	gradingResult, assignment, err := fetchSubmission(fullSubmissionID, submissionDir)
	if err != nil {
		return nil, err
	}

	analysis := &model.IndividualAnalysis{
		AnalysisTimestamp: timestamp.Now(),
		Options:           assignment.AssignmentAnalysisOptions,

		FullID:       gradingResult.Info.ID,
		ShortID:      gradingResult.Info.ShortID,
		CourseID:     gradingResult.Info.CourseID,
		AssignmentID: gradingResult.Info.AssignmentID,
		UserEmail:    gradingResult.Info.User,

		SubmissionStartTime: gradingResult.Info.GradingStartTime,
		Score:               gradingResult.Info.Score,
	}

	// TEST
	fmt.Printf("TEST - %s - computeSingleIndividualAnalysis start: '%v'.\n", options.InitiatorEmail, submissionDir)

	fileInfos, skipped, loc, err := individualFileAnalysis(submissionDir, assignment)
	if err != nil {
		analysis.Failure = true
		analysis.FailureMessage = fmt.Sprintf("Failed to compute individual analysis for '%s': '%s'.", fullSubmissionID, err.Error())

		return analysis, nil
	}

	// TEST
	fmt.Printf("TEST - %s - computeSingleIndividualAnalysis result: '%v'.\n", options.InitiatorEmail, util.MustToJSON(fileInfos))

	analysis.Files = fileInfos
	analysis.SkippedFiles = skipped
	analysis.LinesOfCode = loc

	if computeDeltas {
		err = computeDelta(options, analysis, assignment, gradingResult)
		if err != nil {
			return nil, err
		}
	}

	return analysis, nil
}

func computeDelta(options AnalysisOptions, analysis *model.IndividualAnalysis, assignment *model.Assignment, gradingResult *model.GradingResult) error {
	previousSubmissionID, err := db.GetPreviousSubmissionID(assignment, gradingResult.Info.User, gradingResult.Info.ShortID)
	if err != nil {
		return fmt.Errorf("Failed to get previous submission for delta computation: '%w'.", err)
	}

	if previousSubmissionID == "" {
		return nil
	}

	previousAnalysis, err := computeSingleIndividualAnalysis(options, previousSubmissionID, false)
	if err != nil {
		return fmt.Errorf("Failed to analyze previous submission for delta computation: '%w'.", err)
	}

	timeDelta := (analysis.SubmissionStartTime - previousAnalysis.SubmissionStartTime)
	timeDeltaHours := timeDelta.ToHours()

	analysis.SubmissionTimeDelta = timeDelta.ToMSecs()
	analysis.LinesOfCodeDelta = (analysis.LinesOfCode - previousAnalysis.LinesOfCode)
	analysis.ScoreDelta = (analysis.Score - previousAnalysis.Score)

	if !util.IsZero(timeDeltaHours) {
		analysis.LinesOfCodeVelocity = float64(analysis.LinesOfCodeDelta) / timeDeltaHours
		analysis.ScoreVelocity = analysis.ScoreDelta / timeDeltaHours
	}

	return nil
}

func individualFileAnalysis(submissionDir string, assignment *model.Assignment) ([]model.AnalysisFileInfo, []string, int, error) {
	// Allow a failure for testing.
	if testFailIndividualAnalysis {
		return nil, nil, 0, fmt.Errorf("Test failure.")
	}

	renames, err := prepSourceFiles(submissionDir)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("Failed to prepare source files: '%w'.", err)
	}

	relpaths, err := util.GetAllDirents(submissionDir, true, true)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("Failed to get files: '%w'.", err)
	}

	totalLOC := 0
	infos := make([]model.AnalysisFileInfo, 0, len(relpaths))
	skipped := make([]string, 0)

	for _, relpath := range relpaths {
		// Check if this file should be skipped because of inclusions/exclusions.
		if (assignment.AssignmentAnalysisOptions != nil) && !assignment.AssignmentAnalysisOptions.MatchRelpath(relpath) {
			skipped = append(skipped, relpath)
			continue
		}

		path := filepath.Join(submissionDir, relpath)
		loc, _, err := util.LinesOfCode(path)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("Unable to count lines of code for '%s': '%w'.", relpath, err)
		}

		info := model.AnalysisFileInfo{
			Filename:         relpath,
			OriginalFilename: renames[relpath],
			LinesOfCode:      loc,
		}

		totalLOC += loc
		infos = append(infos, info)
	}

	return infos, skipped, totalLOC, nil
}

func collectIndividualStats(fullSubmissionIDs []string, totalRunTime int64, initiatorEmail string) {
	collectAnalysisStats(fullSubmissionIDs, totalRunTime, initiatorEmail, "individual")
}
