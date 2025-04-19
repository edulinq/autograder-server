package analysis

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

var testFailIndividualAnalysis bool = false

func IndividualAnalysis(options AnalysisOptions) ([]*model.IndividualAnalysis, int, error) {
	// Sort the ids so the result will be consistently ordered.
	fullSubmissionIDs := slices.Clone(options.ResolvedSubmissionIDs)
	slices.Sort(fullSubmissionIDs)

	if len(fullSubmissionIDs) == 0 {
		return nil, 0, nil
	}

	// Lock based on the first seen course.
	// This is to prevent multiple requests using up all the cores.
	lockCourseID, _, _, _, err := common.SplitFullSubmissionID(fullSubmissionIDs[0])
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to get locking course: '%w'.", err)
	}

	options.LockKey = fmt.Sprintf("analysis-individual-course-%s", lockCourseID)

	options.PoolSize = config.ANALYSIS_INDIVIDUAL_COURSE_POOL_SIZE.Get()

	err = options.JobOptions.Validate()
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to validate job options: '%v'.", err)
	}

	job := jobmanager.Job[string, *model.IndividualAnalysis]{
		JobOptions:        options.JobOptions,
		WorkItems:         fullSubmissionIDs,
		RetrieveFunc:      getCachedIndividualResults,
		RemoveStorageFunc: db.RemoveIndividualAnalysis,
		WorkFunc: func(fullSubmissionID string) (*model.IndividualAnalysis, int64, error) {
			return runSingleIndividualAnalysis(options, fullSubmissionID)
		},
	}

	output, err := job.Run()
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to run individual analysis job: '%v'.", err)
	}

	collectIndividualStats(fullSubmissionIDs, output.RunTime, options.InitiatorEmail)

	return output.ResultItems, len(output.RemainingItems), nil
}

func getCachedIndividualResults(fullSubmissionIDs []string) ([]*model.IndividualAnalysis, []string, error) {
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

func runSingleIndividualAnalysis(options AnalysisOptions, fullSubmissionID string) (*model.IndividualAnalysis, int64, error) {
	// Lock this id so we don't try to do the analysis multiple times.
	lockKey := fmt.Sprintf("analysis-individual-%s", fullSubmissionID)
	lockmanager.Lock(lockKey)
	defer lockmanager.Unlock(lockKey)

	// The context has been canceled while waiting for a lock, abandon this analysis.
	if options.Context.Err() != nil {
		return nil, 0, nil
	}

	startTime := timestamp.Now()

	// Check the DB for a complete analysis.
	if !options.OverwriteRecords {
		result, err := db.GetSingleIndividualAnalysis(fullSubmissionID)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to check DB for cached individual analysis for '%s': '%w'.", fullSubmissionID, err)
		}

		if result != nil {
			return result, 0, nil
		}
	}

	// Nothing cached, compute the analysis.
	result, err := computeSingleIndividualAnalysis(options, fullSubmissionID, true)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to compute individual analysis for '%s': '%w'.", fullSubmissionID, err)
	}

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

	fileInfos, skipped, loc, err := individualFileAnalysis(submissionDir, assignment)
	if err != nil {
		analysis.Failure = true
		analysis.FailureMessage = fmt.Sprintf("Failed to compute individual analysis for '%s': '%s'.", fullSubmissionID, err.Error())

		return analysis, nil
	}

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
