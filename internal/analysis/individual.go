package analysis

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

const MSECS_PER_HOUR float64 = float64(1000.0 * 60.0 * 60.0)

func IndividualAnalysis(fullSubmissionIDs []string, blockForResults bool, initiatorEmail string) ([]*model.IndividualAnalysis, int, error) {
	completeAnalysis, remainingIDs, err := getCachedIndividualResults(fullSubmissionIDs)
	if err != nil {
		return nil, 0, err
	}

	if blockForResults {
		results, err := runIndividualAnalysis(remainingIDs, initiatorEmail)
		if err != nil {
			return nil, 0, err
		}

		completeAnalysis = append(completeAnalysis, results...)
		remainingIDs = nil
	} else {
		go func() {
			_, err := runIndividualAnalysis(remainingIDs, initiatorEmail)
			if err != nil {
				log.Error("Failure during asynchronous individual analysis.", err)
			}
		}()
	}

	return completeAnalysis, len(remainingIDs), nil
}

func getCachedIndividualResults(fullSubmissionIDs []string) ([]*model.IndividualAnalysis, []string, error) {
	// Sort the ids so the result will be consistently ordered.
	fullSubmissionIDs = slices.Clone(fullSubmissionIDs)
	slices.Sort(fullSubmissionIDs)

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

func runIndividualAnalysis(fullSubmissionIDs []string, initiatorEmail string) ([]*model.IndividualAnalysis, error) {
	results := make([]*model.IndividualAnalysis, 0, len(fullSubmissionIDs))
	var errs error = nil
	totalRunTime := int64(0)

	for _, fullSubmissionID := range fullSubmissionIDs {
		result, runTime, err := runSingleIndividualAnalysis(fullSubmissionID)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to perform individual analysis on submission %s: '%w'.", fullSubmissionID, err))
		} else {
			results = append(results, result)
			totalRunTime += runTime
		}
	}

	collectIndividualStats(fullSubmissionIDs, totalRunTime, initiatorEmail)

	return results, errs
}

func runSingleIndividualAnalysis(fullSubmissionID string) (*model.IndividualAnalysis, int64, error) {
	// Lock this id so we don't try to do the analysis multiple times.
	lockKey := fmt.Sprintf("analysis-individual-%s", fullSubmissionID)
	common.Lock(lockKey)
	defer common.Unlock(lockKey)

	startTime := timestamp.Now()

	// Check the DB for a complete analysis.
	result, err := db.GetSingleIndividualAnalysis(fullSubmissionID)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to check DB for cached individual analysis for '%s': '%w'.", fullSubmissionID, err)
	}

	if result != nil {
		return result, 0, nil
	}

	// Nothing cached, compute the analsis.
	result, err = computeSingleIndividualAnalysis(fullSubmissionID, true)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to compute individual analysis for '%s': '%w'.", fullSubmissionID, err)
	}

	// Store the result.
	err = db.StoreIndividualAnalysis([]*model.IndividualAnalysis{result})
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to store individual analysis for '%s' in DB: '%w'.", fullSubmissionID, err)
	}

	runTime := int64(timestamp.Now() - startTime)

	return result, runTime, nil
}

func computeSingleIndividualAnalysis(fullSubmissionID string, computeDeltas bool) (*model.IndividualAnalysis, error) {
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

	fileInfos, loc, err := individualFileAnalysis(submissionDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to compute individual analysis for %v: '%w'.", fullSubmissionID, err)
	}

	analysis := &model.IndividualAnalysis{
		AnalysisTimestamp: timestamp.Now(),

		FullID:       gradingResult.Info.ID,
		ShortID:      gradingResult.Info.ShortID,
		CourseID:     gradingResult.Info.CourseID,
		AssignmentID: gradingResult.Info.AssignmentID,
		UserEmail:    gradingResult.Info.User,

		SubmissionStartTime: gradingResult.Info.GradingStartTime,
		Score:               gradingResult.Info.Score,

		Files:       fileInfos,
		LinesOfCode: loc,
	}

	if computeDeltas {
		err = computeDelta(analysis, assignment, gradingResult)
		if err != nil {
			return nil, err
		}
	}

	return analysis, nil
}

func computeDelta(analysis *model.IndividualAnalysis, assignment *model.Assignment, gradingResult *model.GradingResult) error {
	previousSubmissionID, err := db.GetPreviousSubmissionID(assignment, gradingResult.Info.User, gradingResult.Info.ShortID)
	if err != nil {
		return fmt.Errorf("Failed to get previous submission for delta computation: '%w'.", err)
	}

	if previousSubmissionID == "" {
		return nil
	}

	previousAnalysis, err := computeSingleIndividualAnalysis(previousSubmissionID, false)
	if err != nil {
		return fmt.Errorf("Failed to analyze previous submission for delta computation: '%w'.", err)
	}

	timeDeltaMSecs := float64((analysis.SubmissionStartTime - previousAnalysis.SubmissionStartTime).ToMSecs())
	timeDeltaHours := timeDeltaMSecs / MSECS_PER_HOUR

	analysis.LinesOfCodeDelta = (analysis.LinesOfCode - previousAnalysis.LinesOfCode)
	analysis.ScoreDelta = (analysis.Score - previousAnalysis.Score)

	if !util.IsZero(timeDeltaHours) {
		analysis.LinesOfCodeVelocity = float64(analysis.LinesOfCodeDelta) / timeDeltaHours
		analysis.ScoreVelocity = analysis.ScoreDelta / timeDeltaHours
	}

	return nil
}

func individualFileAnalysis(submissionDir string) ([]model.AnalysisFileInfo, int, error) {
	renames, err := prepSourceFiles(submissionDir)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to prepare source files: '%w'.", err)
	}

	relpaths, err := util.GetAllRelativeFiles(submissionDir)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to get files: '%w'.", err)
	}

	totalLOC := 0
	infos := make([]model.AnalysisFileInfo, 0, len(relpaths))

	for _, relpath := range relpaths {
		path := filepath.Join(submissionDir, relpath)
		loc, _, err := util.LinesOfCode(path)
		if err != nil {
			return nil, 0, fmt.Errorf("Unable to count lines of code for '%s': '%w'.", relpath, err)
		}

		info := model.AnalysisFileInfo{
			Filename:         relpath,
			OriginalFilename: renames[relpath],
			LinesOfCode:      loc,
		}

		totalLOC += loc
		infos = append(infos, info)
	}

	return infos, totalLOC, nil
}

func collectIndividualStats(fullSubmissionIDs []string, totalRunTime int64, initiatorEmail string) {
	collectAnalysisStats(fullSubmissionIDs, totalRunTime, initiatorEmail, "individual")
}
