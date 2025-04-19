package analysis

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type AnalysisOptions struct {
	jobmanager.JobOptions

	// Don't save anything.
	DryRun bool `json:"dry-run"`

	// The raw submission specifications to analyze.
	RawSubmissionSpecs []string `json:"submissions"`

	// Email of the person making the request for logging/stats purposes.
	InitiatorEmail string `json:"-"`

	ResolvedSubmissionIDs []string `json:"-"`
}

// Prepare any source files in a directory for analysis.
// The source files may be changed or moved.
// If a file is moved, then the first return (renames) will map the new relpath to the old relpath.
func prepSourceFiles(inputDir string) (map[string]string, error) {
	inputDir = util.ShouldAbs(inputDir)

	renames := make(map[string]string, 0)

	relpaths, err := util.GetAllDirents(inputDir, true, true)
	if err != nil {
		return nil, fmt.Errorf("Failed to get files from dir: '%w'.", err)
	}

	for _, relpath := range relpaths {
		newPath, err := prepSourceFile(filepath.Join(inputDir, relpath))
		if err != nil {
			return nil, fmt.Errorf("Failed to prepare source file '%s': '%w'.", relpath, err)
		}

		if newPath != "" {
			newRelpath := util.RelPath(newPath, inputDir)
			renames[newRelpath] = relpath
		}
	}

	return renames, nil
}

func prepSourceFile(path string) (string, error) {
	ext := filepath.Ext(path)
	if ext == ".ipynb" {
		newPath := strings.TrimSuffix(path, ".ipynb") + ".py"
		for util.PathExists(newPath) {
			newPath = "_" + newPath
		}

		code, err := util.ExtractPythonCodeFromNotebookFile(path)
		if err != nil {
			return "", fmt.Errorf("Unable to extract Python notebook code: '%w'.", err)
		}

		err = util.WriteFile(code, newPath)
		if err != nil {
			return "", fmt.Errorf("Unable to write new Python file from Python notebook source: '%w'.", err)
		}

		err = util.RemoveDirent(path)
		if err != nil {
			return "", fmt.Errorf("Unable to remove old Python notebook file: '%w'.", err)
		}

		return newPath, nil
	}

	return "", nil
}

func fetchSubmission(fullID string, baseDir string) (*model.GradingResult, *model.Assignment, error) {
	courseID, assignmentID, userEmail, shortID, err := common.SplitFullSubmissionID(fullID)
	if err != nil {
		return nil, nil, err
	}

	assignment, err := db.GetAssignment(courseID, assignmentID)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to fetch assignment %s.%s: '%w'.", courseID, assignmentID, err)
	}

	gradingResult, err := db.GetSubmissionContents(assignment, userEmail, shortID)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to fetch submission contents for '%s': '%w'.", fullID, err)
	}

	if gradingResult == nil {
		return nil, nil, fmt.Errorf("Could not find submission '%s'.", fullID)
	}

	err = util.GzipBytesToDirectory(baseDir, gradingResult.InputFilesGZip)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to write submission input to temp dir: '%w'.", err)
	}

	return gradingResult, assignment, nil
}

// Store stats on how long the analysis took.
// All represented courses will get the same time logged.
// If only one assignment is present for a course it will be used in the metric,
// otherwise no assignment will be used.
func collectAnalysisStats(fullSubmissionIDs []string, totalRunTime int64, initiatorEmail string, analysisType string) {
	if totalRunTime <= 0 {
		return
	}

	// {course: assignment, ...}
	seenIdentifiers := make(map[string]string)

	for _, fullSubmissionID := range fullSubmissionIDs {
		courseID, assignmentID, _, _, err := common.SplitFullSubmissionID(fullSubmissionID)
		if err != nil {
			continue
		}

		assignment, ok := seenIdentifiers[courseID]
		if !ok {
			// This course has not been seen, enter it with the current assignment.
			seenIdentifiers[courseID] = assignmentID
		} else if assignmentID != assignment {
			// This course has been seen before, and has a different assignment.
			// Zero out the assignment (as there is more than one in the keys).
			// With a zero value, it will never match another real assignment.
			seenIdentifiers[courseID] = ""
		}
	}

	if len(seenIdentifiers) == 0 {
		log.Error("Could not find identifiers for analysis stat collection.", log.NewAttr("submission-ids", fullSubmissionIDs))
		return
	}

	now := timestamp.Now()

	for courseID, assignmentID := range seenIdentifiers {
		metric := stats.Metric{
			Timestamp: now,
			Type:      stats.MetricTypeCodeAnalysisTime,
			Value:     float64(totalRunTime),
			Attributes: map[stats.MetricAttribute]any{
				stats.MetricAttributeAnalysisType: analysisType,
				stats.MetricAttributeCourseID:     courseID,
			},
		}

		// Add optional fields if non-empty.
		metric.SetUserEmail(initiatorEmail)
		metric.SetAssignmentID(assignmentID)

		stats.AsyncStoreMetric(&metric)
	}
}
