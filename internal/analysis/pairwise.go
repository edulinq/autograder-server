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
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

var testFailPairwiseAnalysis bool = false

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
func PairwiseAnalysis(options AnalysisOptions) ([]*model.PairwiseAnalysis, int, error) {
	_, err := getEngines()
	if err != nil {
		return nil, 0, err
	}

	allKeys := createPairwiseKeys(options.ResolvedSubmissionIDs)
	if len(allKeys) == 0 {
		return []*model.PairwiseAnalysis{}, 0, nil
	}

	// Lock based on the first seen course.
	// This is to prevent multiple requests using up all the cores.
	lockCourseID, _, _, _, err := common.SplitFullSubmissionID(allKeys[0][0])
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to get locking course: '%w'.", err)
	}

	options.LockKey = fmt.Sprintf("analysis-pairwise-course-%s", lockCourseID)

	options.PoolSize = config.ANALYSIS_PAIRWISE_COURSE_POOL_SIZE.Get()

	err = options.JobOptions.Validate()
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to validate job options: '%v'.", err)
	}

	templateFileStore := NewTemplateFileStore()
	defer templateFileStore.Close()

	job := jobmanager.Job[model.PairwiseKey, *model.PairwiseAnalysis]{
		JobOptions:        options.JobOptions,
		WorkItems:         allKeys,
		RetrieveFunc:      getCachedPairwiseResults,
		RemoveStorageFunc: db.RemovePairwiseAnalysis,
		WorkFunc: func(key model.PairwiseKey) (*model.PairwiseAnalysis, int64, error) {
			return runSinglePairwiseAnalysis(options, key, templateFileStore)
		},
	}

	output, err := job.Run()
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to run pairwise analysis job: '%v'.", err)
	}

	collectPairwiseStats(allKeys, output.RunTime, options.InitiatorEmail)

	return output.ResultItems, len(output.RemainingItems), nil
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

func getCachedPairwiseResults(allKeys []model.PairwiseKey) ([]*model.PairwiseAnalysis, []model.PairwiseKey, error) {
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

func runSinglePairwiseAnalysis(options AnalysisOptions, pairwiseKey model.PairwiseKey, templateFileStore *TemplateFileStore) (*model.PairwiseAnalysis, int64, error) {
	// Lock this key so we don't try to do the analysis multiple times.
	lockKey := fmt.Sprintf("analysis-pairwise-single-%s", pairwiseKey.String())
	lockmanager.Lock(lockKey)
	defer lockmanager.Unlock(lockKey)

	// The context has been canceled while waiting for a lock, abandon this analysis.
	if options.Context.Err() != nil {
		return nil, 0, nil
	}

	// Check the DB for a complete analysis.
	if !options.OverwriteRecords {
		result, err := db.GetSinglePairwiseAnalysis(pairwiseKey)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to check DB for cached pairwise analysis for '%s': '%w'.", pairwiseKey.String(), err)
		}

		if result != nil {
			return result, 0, nil
		}
	}

	// Nothing cached, compute the analsis.
	result, runTime, err := computeSinglePairwiseAnalysis(options, pairwiseKey, templateFileStore)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to compute pairwise analysis for '%s': '%w'.", pairwiseKey.String(), err)
	}

	// Store the result.
	if !options.DryRun && (options.Context.Err() == nil) {
		err = db.StorePairwiseAnalysis([]*model.PairwiseAnalysis{result})
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to store pairwise analysis for '%s' in DB: '%w'.", pairwiseKey.String(), err)
		}
	}

	return result, runTime, nil
}

func computeSinglePairwiseAnalysis(options AnalysisOptions, pairwiseKey model.PairwiseKey, templateFileStore *TemplateFileStore) (*model.PairwiseAnalysis, int64, error) {
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

	fileSimilarities, unmatches, skipped, totalRunTime, err := computeFileSims(options, submissionDirs, optionsAssignment, templateFileStore)
	if err != nil {
		message := fmt.Sprintf("Failed to compute similarities for %v: '%s'.", pairwiseKey, err.Error())
		analysis := model.NewFailedPairwiseAnalysis(pairwiseKey, optionsAssignment, message)
		return analysis, 0, nil
	}

	if options.Context.Err() != nil {
		return nil, 0, nil
	}

	analysis := model.NewPairwiseAnalysis(pairwiseKey, optionsAssignment, fileSimilarities, unmatches, skipped)

	return analysis, totalRunTime, nil
}

func computeFileSims(options AnalysisOptions, inputDirs [2]string, assignment *model.Assignment, templateFileStore *TemplateFileStore) (map[string][]*model.FileSimilarity, [][2]string, []string, int64, error) {
	// Allow a failure for testing.
	if testFailPairwiseAnalysis {
		return nil, nil, nil, 0, fmt.Errorf("Test failure.")
	}

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

				similarity, runTime, err := simEngine.ComputeFileSimilarity(paths, templatePath, options.Context)
				if err != nil {
					errs[index] = fmt.Errorf("Unable to compute similarity for '%s' using engine '%s': '%w'", relpath, simEngine.GetName(), err)
				} else if similarity != nil {
					similarity.Filename = relpath
					similarity.OriginalFilename = renames[relpath]

					tempSimilarities[index] = similarity
					runTimes[index] = runTime
				}
			}(i, engine)
		}

		// Wait for all engines to complete.
		engineWaitGroup.Wait()

		if options.Context.Err() != nil {
			return nil, nil, nil, 0, nil
		}

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
