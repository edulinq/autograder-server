package disk

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const DISK_DB_ANALYSIS_INDIVIDUAL_FILENAME = "analysis-individual.jsonl"

func (this *backend) GetIndividualAnalysis(fullSubmissionIDs []string) (map[string]*model.IndividualAnalysis, error) {
	this.analysisIndividualLock.RLock()
	defer this.analysisIndividualLock.RUnlock()

	// Craete a common result map to collect data from all relevant files.
	// This map will also serve as a lookup so we can quickly find matching records.
	results := make(map[string]*model.IndividualAnalysis, len(fullSubmissionIDs))

	// Split the queries by course.
	courseQueries := splitSubmissionIDsByCourse(fullSubmissionIDs)

	// Query each course.
	for courseID, ids := range courseQueries {
		err := this.getCourseIndividualAnalysis(courseID, ids, results)
		if err != nil {
			return nil, fmt.Errorf("Failed to query individual analysis for course '%s': '%w'.", courseID, err)
		}
	}

	return results, nil
}

func (this *backend) getCourseIndividualAnalysis(courseID string, fullSubmissionIDs []string, results map[string]*model.IndividualAnalysis) error {
	// Add IDs so they can be used as a lookup.
	for _, fullSubmissionID := range fullSubmissionIDs {
		results[fullSubmissionID] = nil
	}

	applyFunc := func(index int, record *model.IndividualAnalysis, line string) {
		_, ok := results[record.FullID]
		if ok {
			results[record.FullID] = record
		}
	}

	path := this.getAnalysisIndividualPath(courseID)
	err := util.ApplyJSONLFile(path, model.IndividualAnalysis{}, applyFunc)

	// Cleanup missing ids.
	for _, fullSubmissionID := range fullSubmissionIDs {
		if results[fullSubmissionID] == nil {
			delete(results, fullSubmissionID)
		}
	}

	return err
}

func (this *backend) StoreIndividualAnalysis(allRecords []*model.IndividualAnalysis) error {
	this.analysisIndividualLock.Lock()
	defer this.analysisIndividualLock.Unlock()

	courseRecords := make(map[string][]*model.IndividualAnalysis)
	for _, record := range allRecords {
		if record.CourseID == "" {
			// This would be a bit strange, just log and skip it.
			log.Warn("Found empty course ID in individual analysis.", log.NewAttr("record", record))
			continue
		}

		courseRecords[record.CourseID] = append(courseRecords[record.CourseID], record)
	}

	for courseID, records := range courseRecords {
		path := this.getAnalysisIndividualPath(courseID)

		err := util.AppendJSONLFileMany(path, records)
		if err != nil {
			return fmt.Errorf("Failed to store individual analysis for course '%s': '%w'.", courseID, err)
		}
	}

	return nil
}

func (this *backend) RemoveIndividualAnalysis(fullSubmissionIDs []string) error {
	this.analysisIndividualLock.Lock()
	defer this.analysisIndividualLock.Unlock()

	courseQueries := splitSubmissionIDsByCourse(fullSubmissionIDs)

	for courseID, ids := range courseQueries {
		err := this.removeCourseIndividualAnalysis(courseID, ids)
		if err != nil {
			return fmt.Errorf("Failed to remove individual analysis for course '%s': '%w'.", courseID, err)
		}
	}

	return nil
}

func (this *backend) removeCourseIndividualAnalysis(courseID string, fullSubmissionIDs []string) error {
	slices.Sort(fullSubmissionIDs)

	shouldRemoveFunc := func(record *model.IndividualAnalysis) bool {
		_, exists := slices.BinarySearch(fullSubmissionIDs, record.FullID)
		return exists
	}

	path := this.getAnalysisIndividualPath(courseID)
	return util.RemoveEntriesJSONLFile(path, model.IndividualAnalysis{}, shouldRemoveFunc)
}

func splitSubmissionIDsByCourse(fullSubmissionIDs []string) map[string][]string {
	courseQueries := make(map[string][]string)
	for _, fullSubmissionID := range fullSubmissionIDs {
		courseID, _, _, _, err := common.SplitFullSubmissionID(fullSubmissionID)
		if err != nil {
			// This would be a bit strange, just log and skip it.
			log.Warn("Could not split individual analysis query (full submission ID).", err, log.NewAttr("submission-id", fullSubmissionID))
			continue
		}

		// Add the ID to the course-specific query.
		courseQueries[courseID] = append(courseQueries[courseID], fullSubmissionID)
	}

	return courseQueries
}

func (this *backend) getAnalysisIndividualPath(courseID string) string {
	return filepath.Join(this.getCourseDirFromID(courseID), DISK_DB_ANALYSIS_INDIVIDUAL_FILENAME)
}
