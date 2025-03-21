package analysis

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sync"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/analysis/dolos"
	"github.com/edulinq/autograder/internal/analysis/jplag"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

var defaultSimilarityEngines []core.SimilarityEngine = []core.SimilarityEngine{
	dolos.GetEngine(),
	jplag.GetEngine(),
}

// Unit testing will use a fake engine.
// This will force testing to use the real engines.
var forceDefaultEnginesForTesting bool = false

// Perform a pairwise analysis on a list of full submission IDs.
// Note that these submissions could technically be from different courses/assignments.
// All non-identity pairs will be analyzed, with the lexicographically lower submission ID being on the LHS.
// The LHS's course will be used for contextual locking,
// so only one instance of analysis can be performed at the same time for each (course, tool) pair.
// All courses present in the analysis will have runtime stats logged for it.
// The passed in email should be the user that requested this analysis, stats will be logged with that email.
// Results will be saved to the database for use in future calls.
// If only some results are cached,
// then those will be fetched from the database while the rest are computed.
// If options.WaitForCompletion is true, then this function will block until all requested results are computed.
// Otherwise, this function will return any cached results from the database
// and the remaining analysis will be done asynchronously.
// Returns: (complete results, number of pending analysis runs, error)
func PairwiseAnalysis(options AnalysisOptions, initiatorEmail string) ([]*model.PairwiseAnalysis, int, error) {
	_, err := getEngines()
	if err != nil {
		return nil, 0, err
	}

	completeAnalysis, remainingKeys, err := getCachedPairwiseResults(options)
	if err != nil {
		return nil, 0, err
	}

	if options.WaitForCompletion {
		results, err := runPairwiseAnalysis(options, remainingKeys, initiatorEmail)
		if err != nil {
			return nil, 0, err
		}

		completeAnalysis = append(completeAnalysis, results...)
		remainingKeys = nil
	} else {
		go func() {
			_, err := runPairwiseAnalysis(options, remainingKeys, initiatorEmail)
			if err != nil {
				log.Error("Failure during asynchronous pairwise analysis.", err)
			}
		}()
	}

	return completeAnalysis, len(remainingKeys), nil
}

func createPairwiseKeys(fullSubmissionIDs []string) []model.PairwiseKey {
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

	return allKeys
}

func getCachedPairwiseResults(options AnalysisOptions) ([]*model.PairwiseAnalysis, []model.PairwiseKey, error) {
	allKeys := createPairwiseKeys(options.ResolvedSubmissionIDs)

	// If we are overwriting the cache, don't query the DB for any of the cached results.
	if options.OverwriteCache {
		return make([]*model.PairwiseAnalysis, 0), allKeys, nil
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

// Lock based on course and then run the analysis in a parallel pool.
func runPairwiseAnalysis(options AnalysisOptions, keys []model.PairwiseKey, initiatorEmail string) ([]*model.PairwiseAnalysis, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	// Lock based on the first seen course.
	// This is to prevent multiple requests using up all the cores.
	lockCourseID, _, _, _, err := common.SplitFullSubmissionID(keys[0][0])
	if err != nil {
		return nil, fmt.Errorf("Unable to get locking course: '%w'.", err)
	}

	lockKey := fmt.Sprintf("analysis-pairwise-course-%s", lockCourseID)
	noLockWait := lockmanager.Lock(lockKey)
	defer lockmanager.Unlock(lockKey)

	// If we had to wait for the lock, then check again for cached results.
	// If there are multiple requests queued up,
	// it will be faster to do a bulk check for cached results instead of checking each one individually.
	if !noLockWait {
		_, keys, err = getCachedPairwiseResults(options)
		if err != nil {
			return nil, fmt.Errorf("Failed to re-check result cache before run: '%w'.", err)
		}
	}

	// If we are overwriting the cache, then remove all the old entries.
	if options.OverwriteCache && !options.DryRun {
		err := db.RemovePairwiseAnalysis(keys)
		if err != nil {
			fmt.Errorf("Failed to remove old individual analysis cache entries: '%w'.", err)
		}
	}

	templateFileStore := NewTemplateFileStore()
	defer templateFileStore.Close()

	poolSize := config.ANALYSIS_PAIRWISE_COURSE_POOL_SIZE.Get()
	type PoolResult struct {
		Result  *model.PairwiseAnalysis
		RunTime int64
		Error   error
	}

	poolResults, _, err := util.RunParallelPoolMap(poolSize, keys, func(key model.PairwiseKey) (PoolResult, error) {
		result, runTime, err := runSinglePairwiseAnalysis(options, key, templateFileStore)
		if err != nil {
			err = fmt.Errorf("Failed to perform pairwise analysis on submissions %s: '%w'.", key.String(), err)
		}

		return PoolResult{result, runTime, err}, nil
	})

	results := make([]*model.PairwiseAnalysis, 0, len(keys))
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

	collectPairwiseStats(keys, totalRunTime, initiatorEmail)

	return results, errs
}

func runSinglePairwiseAnalysis(options AnalysisOptions, pairwiseKey model.PairwiseKey, templateFileStore *TemplateFileStore) (*model.PairwiseAnalysis, int64, error) {
	// Lock this key so we don't try to do the analysis multiple times.
	lockKey := fmt.Sprintf("analysis-pairwise-single-%s", pairwiseKey.String())
	lockmanager.Lock(lockKey)
	defer lockmanager.Unlock(lockKey)

	// Check the DB for a complete analysis.
	if !options.OverwriteCache {
		result, err := db.GetSinglePairwiseAnalysis(pairwiseKey)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to check DB for cached pairwise analysis for '%s': '%w'.", pairwiseKey.String(), err)
		}

		if result != nil {
			return result, 0, nil
		}
	}

	// Nothing cached, compute the analsis.
	result, runTime, err := computeSinglePairwiseAnalysis(pairwiseKey, templateFileStore)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to compute pairwise analysis for '%s': '%w'.", pairwiseKey.String(), err)
	}

	// Store the result.
	if !options.DryRun {
		err = db.StorePairwiseAnalysis([]*model.PairwiseAnalysis{result})
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to store pairwise analysis for '%s' in DB: '%w'.", pairwiseKey.String(), err)
		}
	}

	return result, runTime, nil
}

func computeSinglePairwiseAnalysis(pairwiseKey model.PairwiseKey, templateFileStore *TemplateFileStore) (*model.PairwiseAnalysis, int64, error) {
	tempDir, err := util.MkDirTemp("pairwise-analysis-")
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to make temp dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	var optionsAssignment *model.Assignment = nil

	// Collect both submissions in a temp dir.
	var submissionDirs [2]string
	for i, fullSubmissionID := range pairwiseKey {
		submissionDir := filepath.Join(tempDir, fullSubmissionID)

		_, assignment, err := fetchSubmission(fullSubmissionID, submissionDir)
		if err != nil {
			return nil, 0, err
		}

		if optionsAssignment == nil {
			optionsAssignment = assignment
		}

		submissionDirs[i] = submissionDir
	}

	fileSimilarities, unmatches, skipped, totalRunTime, err := computeFileSims(submissionDirs, optionsAssignment, templateFileStore)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to compute similarities for %v: '%w'.", pairwiseKey, err)
	}

	analysis := model.NewPairwiseAnalysis(pairwiseKey, optionsAssignment, fileSimilarities, unmatches, skipped)

	return analysis, totalRunTime, nil
}

func computeFileSims(inputDirs [2]string, assignment *model.Assignment, templateFileStore *TemplateFileStore) (map[string][]*model.FileSimilarity, [][2]string, []string, int64, error) {
	engines, err := getEngines()
	if err != nil {
		return nil, nil, nil, 0, err
	}

	templateDir := ""
	if templateFileStore != nil {
		templateDir, err = templateFileStore.GetTemplatePath(assignment)
		if err != nil {
			return nil, nil, nil, 0, fmt.Errorf("Failed to get template files: '%w'.", err)
		}
	}

	// When preparing source code, we may rename files (e.g. for iPython notebooks).
	// {newRelpath: oldRelpath, ...}
	renames := make(map[string]string, 0)
	for _, inputDir := range inputDirs {
		partialRenames, err := prepSourceFiles(inputDir)
		if err != nil {
			return nil, nil, nil, 0, fmt.Errorf("Failed to prepare source files: '%w'.", err)
		}

		// Note that we may be overriding some paths, but they shuold have the same information.
		for newRelpath, oldRelpath := range partialRenames {
			renames[newRelpath] = oldRelpath
		}
	}

	// Figure out what files need to be analyzed.
	matches, unmatches, err := util.MatchFiles(inputDirs)
	if err != nil {
		return nil, nil, nil, 0, fmt.Errorf("Failed to find matching files: '%w'.", err)
	}

	totalRunTime := int64(0)
	similarities := make(map[string][]*model.FileSimilarity, len(matches))
	skipped := make([]string, 0)

	for _, relpath := range matches {
		// Check if this file should be skipped because of inclusions/exclusions.
		if (assignment != nil) && (assignment.AssignmentAnalysisOptions != nil) && !assignment.AssignmentAnalysisOptions.MatchRelpath(relpath) {
			skipped = append(skipped, relpath)
			continue
		}

		// Check for the template file.
		templatePath := ""
		if templateDir != "" {
			path := filepath.Join(templateDir, relpath)
			if util.IsFile(path) {
				templatePath = path
			}
		}

		paths := [2]string{
			filepath.Join(inputDirs[0], relpath),
			filepath.Join(inputDirs[1], relpath),
		}

		tempSimilarities := make([]*model.FileSimilarity, len(engines))
		runTimes := make([]int64, len(engines))
		errs := make([]error, len(engines))

		var engineWaitGroup sync.WaitGroup

		for i, engine := range engines {
			// Compute the file similarity for each engine in parallel.
			// Note that because we know the index for each engine up-front, we don't need a channel.
			engineWaitGroup.Add(1)
			go func(index int, simEngine core.SimilarityEngine) {
				defer engineWaitGroup.Done()

				similarity, runTime, err := simEngine.ComputeFileSimilarity(paths, templatePath)
				if err != nil {
					errs[index] = fmt.Errorf("Unable to compute similarity for '%s' using engine '%s': '%w'", relpath, simEngine.GetName(), err)
				} else {
					similarity.Filename = relpath
					similarity.OriginalFilename = renames[relpath]

					tempSimilarities[index] = similarity
					runTimes[index] = runTime
				}
			}(i, engine)
		}

		// Wait for all engines to complete.
		engineWaitGroup.Wait()

		// Collect all the similarities.
		similarities[relpath] = make([]*model.FileSimilarity, 0, len(engines))
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
				return nil, nil, nil, 0, err
			}
		}

		// Sum the run times.
		for _, runTime := range runTimes {
			totalRunTime += runTime
		}
	}

	return similarities, unmatches, skipped, totalRunTime, nil
}

func getEngines() ([]core.SimilarityEngine, error) {
	if !forceDefaultEnginesForTesting && config.UNIT_TESTING_MODE.Get() {
		return []core.SimilarityEngine{&fakeSimiliartyEngine{}}, nil
	}

	engines := make([]core.SimilarityEngine, 0, len(defaultSimilarityEngines))
	for _, engine := range defaultSimilarityEngines {
		if engine.IsAvailable() {
			engines = append(engines, engine)
		}
	}

	if len(engines) == 0 {
		return nil, fmt.Errorf("No similarity engines are currently available (are you using docker?).")
	}

	return engines, nil
}

func collectPairwiseStats(keys []model.PairwiseKey, totalRunTime int64, initiatorEmail string) {
	if totalRunTime <= 0 {
		return
	}

	fullSubmissionIDs := make([]string, 0, len(keys)*2)
	for _, keyPair := range keys {
		for _, key := range keyPair {
			fullSubmissionIDs = append(fullSubmissionIDs, key)
		}
	}

	collectAnalysisStats(fullSubmissionIDs, totalRunTime, initiatorEmail, "pairwise")
}
