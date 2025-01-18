package analysis

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// TEST - Could it be worth is to do a secondary caching at the sim tool level?
//        Then, we could actually just do pairwise checks.
//        This actually just sounds like the normal cache (in db) we were thinking about.

// TEST - Check available engines first.

// TEST - Hard error early if no Docker? Or just, if no engines.

// TEST - Check Cache in DB

// TEST - Save

// TEST
var similarityEngines []core.SimilarityEngine

// Perform a pairwise analysis on a list of full submission IDs.
// Note that these submissions could technically be from different courses/assignments.
// All non-identity pairs will be analyzed, with the lexicographically lower submission ID being on the LHS.
// The LHS's course will be used for contextual locking,
// so only one instance of analysis can be performed at the same time for each (course, tool) pair.
// All courses present in the analysis will have runtime stats logged for it.
// Results will be saved to the database for use in future calls.
// If only some results are cached,
// then those will be fetched from the database while the rest are computed.
func PairwiseAnalysis(fullSubmissionIDs []string) ([]*model.PairWiseAnalysis, error) {
	// TEST - Stats
	// TEST - Locks

	results := []*model.PairWiseAnalysis{}
	var errs error = nil

	// Sort the ids so the result will be consistently ordered.
	fullSubmissionIDs = slices.Sorted(slices.Values(fullSubmissionIDs))

	for i := 0; i < len(fullSubmissionIDs); i++ {
		for j := i + 1; j < len(fullSubmissionIDs); j++ {
			if fullSubmissionIDs[i] == fullSubmissionIDs[j] {
				continue
			}

			// Since the ids are already sorted and i < j, we can guarantee this ordering.
			runIDs := [2]string{fullSubmissionIDs[i], fullSubmissionIDs[j]}

			result, err := pairwiseAnalysis(runIDs)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("Failed to perform pairwise analysis on submissions %v: '%w'.", runIDs, err))
			} else {
				results = append(results, result)
			}
		}
	}

	return results, errs
}

func pairwiseAnalysis(fullSubmissionIDs [2]string) (*model.PairWiseAnalysis, error) {
	tempDir, err := util.MkDirTemp("pairwise-analysis-")
	if err != nil {
		return nil, fmt.Errorf("Failed to make temp dir: '%w'.", err)
	}

	// Collect both submissions in a temp dir.
	var submissionDirs [2]string
	for i, fullSubmissionID := range fullSubmissionIDs {
		submissionDir := filepath.Join(tempDir, fullSubmissionID)

		err = fetchSubmission(fullSubmissionID, submissionDir)
		if err != nil {
			return nil, err
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
		similarities[relpath] = make([]*model.FileSimilarity, 0, len(similarityEngines))

		for _, engine := range similarityEngines {
			similarity, err := engine.ComputeFileSimilarity([2]string{
				filepath.Join(submissionDirs[0], relpath),
				filepath.Join(submissionDirs[1], relpath),
			})
			if err != nil {
				return nil, fmt.Errorf("Unable to compute similarity for '%s' using engine '%s': '%w'", relpath, engine.GetName(), err)
			}

			similarities[relpath] = append(similarities[relpath], similarity)
		}
	}

	analysis := model.PairWiseAnalysis{
		AnalysisTimestamp: timestamp.Now(),
		SubmissionIDs:     fullSubmissionIDs,
		Similarities:      similarities,
		UnmatchedFiles:    unmatches,
	}

	return &analysis, nil
}

func fetchSubmission(fullID string, baseDir string) error {
	courseID, assignmentID, userEmail, shortID, err := common.SplitFullSubmissionID(fullID)
	if err != nil {
		return err
	}

	assignment, err := db.GetAssignment(courseID, assignmentID)
	if err != nil {
		return fmt.Errorf("Failed to fetch assignment %s.%s: '%w'.", courseID, assignmentID, err)
	}

	gradingResult, err := db.GetSubmissionContents(assignment, userEmail, shortID)
	if err != nil {
		return fmt.Errorf("Failed to fetch submission contents for '%s': '%w'.", fullID, err)
	}

	err = util.GzipBytesToDirectory(baseDir, gradingResult.InputFilesGZip)
	if err != nil {
		return fmt.Errorf("Failed to write submission input to temp dir: '%w'.", err)
	}

	return nil
}
