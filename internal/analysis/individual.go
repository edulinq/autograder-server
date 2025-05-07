package analysis

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

var testFailIndividualAnalysis bool = false

func IndividualAnalysis(options AnalysisOptions) (map[string]*model.IndividualAnalysis, int, error) {
	// Sort the ids so the locking key will be consistent.
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

	if !options.RetainOriginalContext && !options.WaitForCompletion {
		options.Context = context.Background()
	}

	job := jobmanager.Job[string, *model.IndividualAnalysis]{
		JobOptions:              &options.JobOptions,
		LockKey:                 fmt.Sprintf("analysis-individual-course-%s", lockCourseID),
		PoolSize:                config.ANALYSIS_INDIVIDUAL_COURSE_POOL_SIZE.Get(),
		ReturnIncompleteResults: !options.WaitForCompletion,
		WorkItems:               fullSubmissionIDs,
		RetrieveFunc:            db.GetIndividualAnalysis,
		StoreFunc:               db.StoreIndividualAnalysis,
		RemoveFunc:              db.RemoveIndividualAnalysis,
		WorkFunc: func(fullSubmissionID string) (*model.IndividualAnalysis, error) {
			return computeSingleIndividualAnalysis(options, fullSubmissionID, true)
		},
		WorkItemKeyFunc: func(fullSubmissionID string) string {
			return fmt.Sprintf("analysis-individual-%s", fullSubmissionID)
		},
		OnSuccess: func(result jobmanager.JobOutput[string, *model.IndividualAnalysis]) {
			collectIndividualStats(fullSubmissionIDs, result.RunTime, options.InitiatorEmail)
		},
	}

	err = job.Validate()
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to validate job: '%v'.", err)
	}

	output := job.Run()
	if output.Error != nil {
		return nil, 0, fmt.Errorf("Failed to run individual analysis job with '%d' work errors: '%v'.", len(output.WorkErrors), output.Error)
	}

	return output.ResultItems, len(output.RemainingItems), nil
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
