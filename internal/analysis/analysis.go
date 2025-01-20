package analysis

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sync"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/analysis/dolos"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// TEST - Stats

// TEST - Assignment override (e.g. language)

var similarityEngines []core.SimilarityEngine = []core.SimilarityEngine{
	dolos.GetEngine(),
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
func PairwiseAnalysis(fullSubmissionIDs []string, blockForResults bool) ([]*model.PairWiseAnalysis, int, error) {
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

func getCachedResults(fullSubmissionIDs []string) ([]*model.PairWiseAnalysis, []model.PairwiseKey, error) {
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
	completeAnalysis := make([]*model.PairWiseAnalysis, 0, len(dbResults))
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

func runPairwiseAnalysis(keys []model.PairwiseKey) ([]*model.PairWiseAnalysis, error) {
	results := make([]*model.PairWiseAnalysis, 0, len(keys))
	var errs error = nil

	for _, key := range keys {
		result, err := runSinglePairwiseAnalysis(key)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to perform pairwise analysis on submissions %s: '%w'.", key.String(), err))
		} else {
			results = append(results, result)
		}
	}

	return results, errs
}

func runSinglePairwiseAnalysis(pairwiseKey model.PairwiseKey) (*model.PairWiseAnalysis, error) {
	// Lock this key so we don't try to do the analysis multiple times.
	lockKey := fmt.Sprintf("analysis-pairwise-%s", pairwiseKey.String())
	common.Lock(lockKey)
	defer common.Unlock(lockKey)

	// Check the DB for a complete analysis.
	result, err := db.GetSinglePairwiseAnalysis(pairwiseKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to check DB for cached pairwise analysis for '%s': '%w'.", pairwiseKey.String(), err)
	}

	if result != nil {
		return result, nil
	}

	// Nothing cached, compute the analsis.
	result, err = computeSinglePairwiseAnalysis(pairwiseKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to compute pairwise analysis for '%s': '%w'.", pairwiseKey.String(), err)
	}

	// Store the result.
	err = db.StorePairwiseAnalysis([]*model.PairWiseAnalysis{result})
	if err != nil {
		return nil, fmt.Errorf("Failed to store pairwise analysis for '%s' in DB: '%w'.", pairwiseKey.String(), err)
	}

	return result, nil
}

func computeSinglePairwiseAnalysis(pairwiseKey model.PairwiseKey) (*model.PairWiseAnalysis, error) {
	tempDir, err := util.MkDirTemp("pairwise-analysis-")
	if err != nil {
		return nil, fmt.Errorf("Failed to make temp dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	lockCourseID := ""

	// Collect both submissions in a temp dir.
	var submissionDirs [2]string
	for i, fullSubmissionID := range pairwiseKey {
		submissionDir := filepath.Join(tempDir, fullSubmissionID)

		courseID, err := fetchSubmission(fullSubmissionID, submissionDir)
		if err != nil {
			return nil, err
		}

		if lockCourseID == "" {
			lockCourseID = courseID
		}

		submissionDirs[i] = submissionDir
	}

	// Figure out what files need to be analyzed.
	matches, unmatches, err := util.MatchFiles(submissionDirs)
	if err != nil {
		return nil, fmt.Errorf("Failed to find matching files: '%w'.", err)
	}

	similarities := make(map[string][]*model.FileSimilarity, len(matches))
	for _, relpath := range matches {
		paths := [2]string{
			filepath.Join(submissionDirs[0], relpath),
			filepath.Join(submissionDirs[1], relpath),
		}

		similarities[relpath] = make([]*model.FileSimilarity, len(similarityEngines))
		errs := make([]error, len(similarityEngines))
		var engineWaitGroup sync.WaitGroup

		for i, engine := range similarityEngines {
			// Compute the file similarity for each engine in parallel.
			// Note that because we know the index for each engine up-front, we don't need a channel.
			engineWaitGroup.Add(1)
			go func(index int, simEngine core.SimilarityEngine) {
				defer engineWaitGroup.Done()

				similarity, err := simEngine.ComputeFileSimilarity(paths, lockCourseID)
				if err != nil {
					errs[index] = fmt.Errorf("Unable to compute similarity for '%s' using engine '%s': '%w'", relpath, simEngine.GetName(), err)
				} else {
					similarities[relpath][index] = similarity
				}
			}(i, engine)
		}

		// Wait for all engines to complete.
		engineWaitGroup.Wait()

		// Check for errors from the children.
		err := errors.Join(errs...)
		if err != nil {
			return nil, err
		}
	}

	analysis := model.PairWiseAnalysis{
		AnalysisTimestamp: timestamp.Now(),
		SubmissionIDs:     pairwiseKey,
		Similarities:      similarities,
		UnmatchedFiles:    unmatches,
	}

	return &analysis, nil
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
