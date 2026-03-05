package disk

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const DISK_DB_ANALYSIS_PRIVATIZED_FILENAME = "analysis-privatized.jsonl"

func (this *backend) GetPrivatizedAnalysis(fullSubmissionIDs []string) (map[string]*model.PrivatizedAnalysis, error) {
	this.analysisPrivatizedLock.RLock()
	defer this.analysisPrivatizedLock.RUnlock()

	results := make(map[string]*model.PrivatizedAnalysis, len(fullSubmissionIDs))

	courseQueries := splitSubmissionIDsByCourse(fullSubmissionIDs)

	for courseID, ids := range courseQueries {
		err := this.getCoursePrivatizedAnalysis(courseID, ids, results)
		if err != nil {
			return nil, fmt.Errorf("Failed to query privatized analysis for course '%s': '%w'.", courseID, err)
		}
	}

	return results, nil
}

func (this *backend) getCoursePrivatizedAnalysis(courseID string, fullSubmissionIDs []string, results map[string]*model.PrivatizedAnalysis) error {
	for _, id := range fullSubmissionIDs {
		results[id] = nil
	}

	applyFunc := func(index int, record *model.PrivatizedAnalysis, line string) {
		if _, ok := results[record.FullSubmissionID]; ok {
			results[record.FullSubmissionID] = record
		}
	}

	path := this.getAnalysisPrivatizedPath(courseID)
	err := util.ApplyJSONLFile(path, model.PrivatizedAnalysis{}, applyFunc)

	for _, id := range fullSubmissionIDs {
		if results[id] == nil {
			delete(results, id)
		}
	}

	return err
}

func (this *backend) StorePrivatizedAnalysis(records []*model.PrivatizedAnalysis) error {
	this.analysisPrivatizedLock.Lock()
	defer this.analysisPrivatizedLock.Unlock()

	courseRecords := make(map[string][]*model.PrivatizedAnalysis)
	for _, record := range records {
		if record.CourseID == "" {
			log.Warn("Found empty course ID in privatized analysis record.", log.NewAttr("submission-id", record.FullSubmissionID))
			continue
		}

		courseRecords[record.CourseID] = append(courseRecords[record.CourseID], record)
	}

	for courseID, batch := range courseRecords {
		path := this.getAnalysisPrivatizedPath(courseID)

		err := util.AppendJSONLFileMany(path, batch)
		if err != nil {
			return fmt.Errorf("Failed to store privatized analysis for course '%s': '%w'.", courseID, err)
		}
	}

	return nil
}

func (this *backend) RemovePrivatizedAnalysis(fullSubmissionIDs []string) error {
	this.analysisPrivatizedLock.Lock()
	defer this.analysisPrivatizedLock.Unlock()

	courseQueries := splitSubmissionIDsByCourse(fullSubmissionIDs)

	for courseID, ids := range courseQueries {
		err := this.removeCoursePrivatizedAnalysis(courseID, ids)
		if err != nil {
			return fmt.Errorf("Failed to remove privatized analysis for course '%s': '%w'.", courseID, err)
		}
	}

	return nil
}

func (this *backend) removeCoursePrivatizedAnalysis(courseID string, fullSubmissionIDs []string) error {
	slices.Sort(fullSubmissionIDs)

	shouldRemove := func(record *model.PrivatizedAnalysis) bool {
		_, exists := slices.BinarySearch(fullSubmissionIDs, record.FullSubmissionID)
		return exists
	}

	path := this.getAnalysisPrivatizedPath(courseID)
	return util.RemoveEntriesJSONLFile(path, model.PrivatizedAnalysis{}, shouldRemove)
}

func (this *backend) getAnalysisPrivatizedPath(courseID string) string {
	return filepath.Join(this.getCourseDirFromID(courseID), DISK_DB_ANALYSIS_PRIVATIZED_FILENAME)
}
