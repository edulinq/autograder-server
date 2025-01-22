package analysis

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/analysis/dolos"
	"github.com/edulinq/autograder/internal/analysis/jplag"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

var similarityEngines []core.SimilarityEngine = []core.SimilarityEngine{
	dolos.GetEngine(),
	jplag.GetEngine(),
}

// Perform a pairwise analysis on a list of full submission IDs.
// Note that these submissions could technically be from different courses/assignments.
// All non-identity pairs will be analyzed, with the lexicographically lower submission ID being on the LHS.
// The LHS's course will be used for contextual locking,
// so only one instance of analysis can be performed at the same time for each (course, tool) pair.
// All courses present in the analysis will have runtime stats logged for it.
// Results will be saved to the database for use in future calls.
// If only some results are cached,
// then those will be fetched from the database while the rest are computed.
// If blockForResults is true, then this function will block until all requested results are computed.
// Otherwise, this function will return any cached results from the database
// and the remaining analysis will be done asynchronously.
// Returns: (complete results, number of pending analysis runs, error)
func PairwiseAnalysis(fullSubmissionIDs []string, blockForResults bool) ([]*model.PairwiseAnalysis, int, error) {
	err := checkEngines()
	if err != nil {
		return nil, 0, err
	}

	completeAnalysis, remainingKeys, err := getCachedResults(fullSubmissionIDs)
	if err != nil {
		return nil, 0, err
	}

	if blockForResults {
		results, err := runPairwiseAnalysis(remainingKeys)
		if err != nil {
			return nil, 0, err
		}

		completeAnalysis = append(completeAnalysis, results...)
		remainingKeys = nil
	} else {
		go func() {
			_, err := runPairwiseAnalysis(remainingKeys)
			if err != nil {
				log.Error("Failure during asynchronous pairwise analysis.", err)
			}
		}()
	}

	return completeAnalysis, len(remainingKeys), nil
}

func getCachedResults(fullSubmissionIDs []string) ([]*model.PairwiseAnalysis, []model.PairwiseKey, error) {
	// Sort the ids so the result will be consistently ordered.
	fullSubmissionIDs = slices.Clone(fullSubmissionIDs)
	slices.Sort(fullSubmissionIDs)

	allKeys := make([]model.PairwiseKey, 0, (len(fullSubmissionIDs) * (len(fullSubmissionIDs) - 1) / 2))
	for i := 0; i < len(fullSubmissionIDs); i++ {
		for j := i + 1; j < len(fullSubmissionIDs); j++ {
			if fullSubmissionIDs[i] == fullSubmissionIDs[j] {
				continue
			}

			allKeys = append(allKeys, model.NewPairwiseKey(fullSubmissionIDs[i], fullSubmissionIDs[j]))
		}
	}

	// Get any already done analysis results from the DB.
	dbResults, err := db.GetPairwiseAnalysis(allKeys)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get cached pairwise analysis from DB: '%w'.", err)
	}

	// Split up the keys into complete and remaining.
	completeAnalysis := make([]*model.PairwiseAnalysis, 0, len(dbResults))
	remainingKeys := make([]model.PairwiseKey, 0, len(allKeys)-len(dbResults))

	for _, key := range allKeys {
		result, ok := dbResults[key]
		if ok {
			completeAnalysis = append(completeAnalysis, result)
		} else {
			remainingKeys = append(remainingKeys, key)
		}
	}

	return completeAnalysis, remainingKeys, nil
}

func runPairwiseAnalysis(keys []model.PairwiseKey) ([]*model.PairwiseAnalysis, error) {
	results := make([]*model.PairwiseAnalysis, 0, len(keys))
	var errs error = nil
	totalRunTime := int64(0)

	for _, key := range keys {
		result, runTime, err := runSinglePairwiseAnalysis(key)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to perform pairwise analysis on submissions %s: '%w'.", key.String(), err))
		} else {
			results = append(results, result)
			totalRunTime += runTime
		}
	}

	collectStats(keys, totalRunTime)

	return results, errs
}

func runSinglePairwiseAnalysis(pairwiseKey model.PairwiseKey) (*model.PairwiseAnalysis, int64, error) {
	// Lock this key so we don't try to do the analysis multiple times.
	lockKey := fmt.Sprintf("analysis-pairwise-%s", pairwiseKey.String())
	common.Lock(lockKey)
	defer common.Unlock(lockKey)

	// Check the DB for a complete analysis.
	result, err := db.GetSinglePairwiseAnalysis(pairwiseKey)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to check DB for cached pairwise analysis for '%s': '%w'.", pairwiseKey.String(), err)
	}

	if result != nil {
		return result, 0, nil
	}

	// Nothing cached, compute the analsis.
	result, runTime, err := computeSinglePairwiseAnalysis(pairwiseKey)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to compute pairwise analysis for '%s': '%w'.", pairwiseKey.String(), err)
	}

	// Store the result.
	err = db.StorePairwiseAnalysis([]*model.PairwiseAnalysis{result})
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to store pairwise analysis for '%s' in DB: '%w'.", pairwiseKey.String(), err)
	}

	return result, runTime, nil
}

func computeSinglePairwiseAnalysis(pairwiseKey model.PairwiseKey) (*model.PairwiseAnalysis, int64, error) {
	tempDir, err := util.MkDirTemp("pairwise-analysis-")
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to make temp dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	lockCourseID := ""

	// Collect both submissions in a temp dir.
	var submissionDirs [2]string
	for i, fullSubmissionID := range pairwiseKey {
		submissionDir := filepath.Join(tempDir, fullSubmissionID)

		courseID, err := fetchSubmission(fullSubmissionID, submissionDir)
		if err != nil {
			return nil, 0, err
		}

		if lockCourseID == "" {
			lockCourseID = courseID
		}

		submissionDirs[i] = submissionDir
	}

	fileSimilarities, unmatches, totalRunTime, err := computeFileSims(submissionDirs, lockCourseID)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to compute similarities for %v: '%w'.", pairwiseKey, err)
	}

	analysis := model.NewPairwiseAnalysis(pairwiseKey, fileSimilarities, unmatches)

	return analysis, totalRunTime, nil
}

// Prepare any source files in a direcory for analysis.
// The source files may be changed or moved.
// If a file is moved, then the first return (renames) will map the new relpath to the old relpath.
func prepSourceFiles(inputDir string) (map[string]string, error) {
	inputDir = util.ShouldAbs(inputDir)

	renames := make(map[string]string, 0)

	relpaths, err := util.GetAllRelativeFiles(inputDir)
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

func computeFileSims(inputDirs [2]string, lockID string) (map[string][]*model.FileSimilarity, [][2]string, int64, error) {
	// When preparing source code, we may rename files (e.g. for iPython notebooks).
	// {newRelpath: oldRelpath, ...}
	renames := make(map[string]string, 0)
	for _, inputDir := range inputDirs {
		partialRenames, err := prepSourceFiles(inputDir)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("Failed to prepare source files: '%w'.", err)
		}

		// Note that we may be overriding some paths, but they shuold have the same information.
		for newRelpath, oldRelpath := range partialRenames {
			renames[newRelpath] = oldRelpath
		}
	}

	// Figure out what files need to be analyzed.
	matches, unmatches, err := util.MatchFiles(inputDirs)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("Failed to find matching files: '%w'.", err)
	}

	totalRunTime := int64(0)
	similarities := make(map[string][]*model.FileSimilarity, len(matches))

	for _, relpath := range matches {
		paths := [2]string{
			filepath.Join(inputDirs[0], relpath),
			filepath.Join(inputDirs[1], relpath),
		}

		tempSimilarities := make([]*model.FileSimilarity, len(similarityEngines))
		runTimes := make([]int64, len(similarityEngines))
		errs := make([]error, len(similarityEngines))

		var engineWaitGroup sync.WaitGroup

		for i, engine := range similarityEngines {
			// Compute the file similarity for each engine in parallel.
			// Note that because we know the index for each engine up-front, we don't need a channel.
			engineWaitGroup.Add(1)
			go func(index int, simEngine core.SimilarityEngine) {
				defer engineWaitGroup.Done()

				similarity, runTime, err := simEngine.ComputeFileSimilarity(paths, lockID)
				if err != nil {
					errs[index] = fmt.Errorf("Unable to compute similarity for '%s' using engine '%s': '%w'", relpath, simEngine.GetName(), err)
				} else {
					similarity.OriginalFilename = renames[relpath]

					tempSimilarities[index] = similarity
					runTimes[index] = runTime
				}
			}(i, engine)
		}

		// Wait for all engines to complete.
		engineWaitGroup.Wait()

		// Collect all the similarities.
		similarities[relpath] = make([]*model.FileSimilarity, 0, len(similarityEngines))
		for _, similarity := range tempSimilarities {
			if similarity != nil {
				similarities[relpath] = append(similarities[relpath], similarity)
			}
		}

		// Check for errors from the children.
		err := errors.Join(errs...)
		if err != nil {
			// If at least one engine worked, don't error out, just log.
			if len(similarities[relpath]) > 0 {
				log.Warn("Not all engines successfully computed similarity. Some engines did complete successfully.", err)
			} else {
				return nil, nil, 0, err
			}
		}

		// Sum the run times.
		for _, runTime := range runTimes {
			totalRunTime += runTime
		}
	}

	return similarities, unmatches, totalRunTime, nil
}

// Store stats on how long the pairwise analysis took.
// All represented courses will get the same time logged.
// If only one assignment is present for a course it will be used in the metric,
// otherwise no assignment will be used.
func collectStats(keys []model.PairwiseKey, totalRunTime int64) {
	// {course: assignment, ...}
	seenIdentifiers := make(map[string]string)

	for _, keyPair := range keys {
		for _, key := range keyPair {
			courseID, assignmentID, _, _, err := common.SplitFullSubmissionID(key)
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
	}

	if len(seenIdentifiers) == 0 {
		log.Error("Could not find identifiers for stat collection of pairwise analysis.", log.NewAttr("keys", keys))
		return
	}

	now := timestamp.Now()

	for courseID, assignmentID := range seenIdentifiers {
		metric := stats.CourseMetric{
			BaseMetric: stats.BaseMetric{
				Timestamp: now,
				Attributes: map[string]any{
					"anslysis-type": "pairwise",
				},
			},
			Type:         stats.CourseMetricTypeCodeAnalysisTime,
			CourseID:     courseID,
			AssignmentID: assignmentID,
			Value:        uint64(totalRunTime),
		}

		stats.AsyncStoreCourseMetric(&metric)
	}
}

func fetchSubmission(fullID string, baseDir string) (string, error) {
	courseID, assignmentID, userEmail, shortID, err := common.SplitFullSubmissionID(fullID)
	if err != nil {
		return "", err
	}

	assignment, err := db.GetAssignment(courseID, assignmentID)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch assignment %s.%s: '%w'.", courseID, assignmentID, err)
	}

	gradingResult, err := db.GetSubmissionContents(assignment, userEmail, shortID)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch submission contents for '%s': '%w'.", fullID, err)
	}

	err = util.GzipBytesToDirectory(baseDir, gradingResult.InputFilesGZip)
	if err != nil {
		return "", fmt.Errorf("Failed to write submission input to temp dir: '%w'.", err)
	}

	return courseID, nil
}

func checkEngines() error {
	available := false
	for _, engine := range similarityEngines {
		available = available || engine.IsAvailable()
	}

	if !available {
		return fmt.Errorf("No similarity engines are currently available (are you using docker?).")
	}

	return nil
}
