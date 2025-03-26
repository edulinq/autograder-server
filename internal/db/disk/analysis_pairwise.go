package disk

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const DISK_DB_ANALYSIS_PAIRWISE_FILENAME = "analysis-pairwise.jsonl"

func (this *backend) GetPairwiseAnalysis(keys []model.PairwiseKey) (map[model.PairwiseKey]*model.PairwiseAnalysis, error) {
	this.analysisPairwiseLock.RLock()
	defer this.analysisPairwiseLock.RUnlock()

	// Craete a common result map to collect data from all relevant files.
	// This map will also serve as a lookup so we can quickly find matching records.
	results := make(map[model.PairwiseKey]*model.PairwiseAnalysis, len(keys))

	// Split the queries by course.
	courseQueries := splitPairwiseKeysByCourse(keys)

	// Query each course.
	for courseID, keys := range courseQueries {
		err := this.getCoursePairwiseAnalysis(courseID, keys, results)
		if err != nil {
			return nil, fmt.Errorf("Failed to query pairwise analysis for course '%s': '%w'.", courseID, err)
		}
	}

	return results, nil
}

func (this *backend) getCoursePairwiseAnalysis(courseID string, keys []model.PairwiseKey, results map[model.PairwiseKey]*model.PairwiseAnalysis) error {
	// Add keys so they can be used as a lookup.
	for _, key := range keys {
		results[key] = nil
	}

	applyFunc := func(index int, record *model.PairwiseAnalysis, line string) {
		_, ok := results[record.SubmissionIDs]
		if ok {
			results[record.SubmissionIDs] = record
		}
	}

	path := this.getAnalysisPairwisePath(courseID)
	err := util.ApplyJSONLFile(path, model.PairwiseAnalysis{}, applyFunc)

	// Cleanup missing keys.
	for _, key := range keys {
		if results[key] == nil {
			delete(results, key)
		}
	}

	return err
}

func (this *backend) StorePairwiseAnalysis(allRecords []*model.PairwiseAnalysis) error {
	this.analysisPairwiseLock.Lock()
	defer this.analysisPairwiseLock.Unlock()

	courseRecords := make(map[string][]*model.PairwiseAnalysis)
	for _, record := range allRecords {
		courseID := record.SubmissionIDs.Course()
		if courseID == "" {
			// This would be a bit strange, just log and skip it.
			log.Warn("Found empty course ID in pairwise analysis.", log.NewAttr("record", record))
			continue
		}

		courseRecords[courseID] = append(courseRecords[courseID], record)
	}

	for courseID, records := range courseRecords {
		path := this.getAnalysisPairwisePath(courseID)

		err := util.AppendJSONLFileMany(path, records)
		if err != nil {
			return fmt.Errorf("Failed to store pairwise analysis for course '%s': '%w'.", courseID, err)
		}
	}

	return nil
}

func (this *backend) RemovePairwiseAnalysis(allKeys []model.PairwiseKey) error {
	this.analysisPairwiseLock.Lock()
	defer this.analysisPairwiseLock.Unlock()

	courseQueries := splitPairwiseKeysByCourse(allKeys)

	for courseID, keys := range courseQueries {
		err := this.removeCoursePairwiseAnalysis(courseID, keys)
		if err != nil {
			return fmt.Errorf("Failed to remove pairwise analysis for course '%s': '%w'.", courseID, err)
		}
	}

	return nil
}

func (this *backend) removeCoursePairwiseAnalysis(courseID string, keys []model.PairwiseKey) error {
	slices.SortFunc(keys, model.ComparePairwiseKey)

	shouldRemoveFunc := func(record *model.PairwiseAnalysis) bool {
		_, exists := slices.BinarySearchFunc(keys, record.SubmissionIDs, model.ComparePairwiseKey)
		return exists
	}

	path := this.getAnalysisPairwisePath(courseID)
	return util.RemoveEntriesJSONLFile(path, model.PairwiseAnalysis{}, shouldRemoveFunc)
}

func splitPairwiseKeysByCourse(keys []model.PairwiseKey) map[string][]model.PairwiseKey {
	courseQueries := make(map[string][]model.PairwiseKey)
	for _, key := range keys {
		courseID := key.Course()
		if courseID == "" {
			// This would be a bit strange, just log and skip it.
			log.Warn("Could not get course from pairwise analysis query (pairwise key).", log.NewAttr("pairwise-key", key))
			continue
		}

		// Add the ID to the course-specific query.
		courseQueries[courseID] = append(courseQueries[courseID], key)
	}

	return courseQueries
}

func (this *backend) getAnalysisPairwisePath(courseID string) string {
	return filepath.Join(this.getCourseDirFromID(courseID), DISK_DB_ANALYSIS_PAIRWISE_FILENAME)
}
