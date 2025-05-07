package analysis

import (
	"context"
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
func PairwiseAnalysis(options AnalysisOptions) (model.PairwiseAnalysisMap, int, error) {
	_, err := getEngines()
	if err != nil {
		return nil, 0, err
	}

	allKeys := createPairwiseKeys(options.ResolvedSubmissionIDs)
	if len(allKeys) == 0 {
		return model.PairwiseAnalysisMap{}, 0, nil
	}

	// Lock based on the first seen course.
	// This is to prevent multiple requests using up all the cores.
	lockCourseID, _, _, _, err := common.SplitFullSubmissionID(allKeys[0][0])
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to get locking course: '%w'.", err)
	}

	if !options.RetainOriginalContext && !options.WaitForCompletion {
		options.Context = context.Background()
	}

	templateFileStore := NewTemplateFileStore()
	defer templateFileStore.Close()

	job := jobmanager.Job[model.PairwiseKey, *model.PairwiseAnalysis]{
		JobOptions:              &options.JobOptions,
		LockKey:                 fmt.Sprintf("analysis-pairwise-course-%s", lockCourseID),
		PoolSize:                config.ANALYSIS_PAIRWISE_COURSE_POOL_SIZE.Get(),
		ReturnIncompleteResults: !options.WaitForCompletion,
		WorkItems:               allKeys,
		RetrieveFunc:            db.GetPairwiseAnalysis,
		StoreFunc:               db.StorePairwiseAnalysis,
		RemoveFunc:              db.RemovePairwiseAnalysis,
		WorkFunc: func(key model.PairwiseKey) (*model.PairwiseAnalysis, error) {
			return computeSinglePairwiseAnalysis(options, key, templateFileStore)
		},
		WorkItemKeyFunc: func(key model.PairwiseKey) string {
			return fmt.Sprintf("analysis-pairwise-single-%s", key.String())
		},
		OnSuccess: func(result jobmanager.JobOutput[model.PairwiseKey, *model.PairwiseAnalysis]) {
			collectPairwiseStats(allKeys, result.RunTime, options.InitiatorEmail)
		},
	}

	err = job.Validate()
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to validate job: '%v'.", err)
	}

	output := job.Run()
	if output.Error != nil {
		return nil, 0, fmt.Errorf("Failed to run pairwise analysis job with '%d' work errors: '%v'.", len(output.WorkErrors), output.Error)
	}

	return output.ResultItems, len(output.RemainingItems), nil
}

func createPairwiseKeys(fullSubmissionIDs []string) []model.PairwiseKey {
	// Sort the ids so the locking key will be consistent.
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

func computeSinglePairwiseAnalysis(options AnalysisOptions, pairwiseKey model.PairwiseKey, templateFileStore *TemplateFileStore) (*model.PairwiseAnalysis, error) {
	tempDir, err := util.MkDirTemp("pairwise-analysis-")
	if err != nil {
		return nil, fmt.Errorf("Failed to make temp dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	var optionsAssignment *model.Assignment = nil

	// Collect both submissions in a temp dir.
	var submissionDirs [2]string
	for i, fullSubmissionID := range pairwiseKey {
		submissionDir := filepath.Join(tempDir, fullSubmissionID)

		_, assignment, err := fetchSubmission(fullSubmissionID, submissionDir)
		if err != nil {
			return nil, err
		}

		if optionsAssignment == nil {
			optionsAssignment = assignment
		}

		submissionDirs[i] = submissionDir
	}

	fileSimilarities, unmatches, skipped, err := computeFileSims(options, submissionDirs, optionsAssignment, templateFileStore)
	if err != nil {
		message := fmt.Sprintf("Failed to compute similarities for %v: '%s'.", pairwiseKey, err.Error())
		analysis := model.NewFailedPairwiseAnalysis(pairwiseKey, optionsAssignment, message)
		return analysis, nil
	}

	if options.Context.Err() != nil {
		return nil, nil
	}

	analysis := model.NewPairwiseAnalysis(pairwiseKey, optionsAssignment, fileSimilarities, unmatches, skipped)

	return analysis, nil
}

func computeFileSims(options AnalysisOptions, inputDirs [2]string, assignment *model.Assignment, templateFileStore *TemplateFileStore) (map[string][]*model.FileSimilarity, [][2]string, []string, error) {
	// Allow a failure for testing.
	if testFailPairwiseAnalysis {
		return nil, nil, nil, fmt.Errorf("Test failure.")
	}

	engines, err := getEngines()
	if err != nil {
		return nil, nil, nil, err
	}

	templateDir := ""
	if templateFileStore != nil {
		templateDir, err = templateFileStore.GetTemplatePath(assignment)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Failed to get template files: '%w'.", err)
		}
	}

	// When preparing source code, we may rename files (e.g. for iPython notebooks).
	// {newRelpath: oldRelpath, ...}
	renames := make(map[string]string, 0)
	for _, inputDir := range inputDirs {
		partialRenames, err := prepSourceFiles(inputDir)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Failed to prepare source files: '%w'.", err)
		}

		// Note that we may be overriding some paths, but they shuold have the same information.
		for newRelpath, oldRelpath := range partialRenames {
			renames[newRelpath] = oldRelpath
		}
	}

	// Figure out what files need to be analyzed.
	matches, unmatches, err := util.MatchFiles(inputDirs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to find matching files: '%w'.", err)
	}

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
		errs := make([]error, len(engines))

		var engineWaitGroup sync.WaitGroup

		for i, engine := range engines {
			// Compute the file similarity for each engine in parallel.
			// Note that because we know the index for each engine up-front, we don't need a channel.
			engineWaitGroup.Add(1)
			go func(index int, simEngine core.SimilarityEngine) {
				defer engineWaitGroup.Done()

				similarity, err := simEngine.ComputeFileSimilarity(paths, templatePath, options.Context)
				if err != nil {
					errs[index] = fmt.Errorf("Unable to compute similarity for '%s' using engine '%s': '%w'", relpath, simEngine.GetName(), err)
				} else if similarity != nil {
					similarity.Filename = relpath
					similarity.OriginalFilename = renames[relpath]

					tempSimilarities[index] = similarity
				}
			}(i, engine)
		}

		// Wait for all engines to complete.
		engineWaitGroup.Wait()

		if options.Context.Err() != nil {
			return nil, nil, nil, nil
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
				return nil, nil, nil, err
			}
		}
	}

	return similarities, unmatches, skipped, nil
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
