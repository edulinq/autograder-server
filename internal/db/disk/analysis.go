package disk

import (
	"path/filepath"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const DISK_DB_ANALYSIS_INDIVIDUAL_FILENAME = "analysis-individual.jsonl"
const DISK_DB_ANALYSIS_PAIRWISE_FILENAME = "analysis-pairwise.jsonl"

func (this *backend) GetIndividualAnalysis(fullSubmissionIDs []string) (map[string]*model.IndividualAnalysis, error) {
	this.analysisIndividualLock.RLock()
	defer this.analysisIndividualLock.RUnlock()

	// Build a lookup that we will also use for storage.
	results := make(map[string]*model.IndividualAnalysis, len(fullSubmissionIDs))
	for _, fullSubmissionID := range fullSubmissionIDs {
		results[fullSubmissionID] = nil
	}

	path := this.getAnalysisIndividualPath()

	err := util.ApplyJSONLFile(path, model.IndividualAnalysis{}, func(index int, record *model.IndividualAnalysis) {
		_, ok := results[record.FullID]
		if ok {
			results[record.FullID] = record
		}
	})

	// Cleanup missing ids.
	for fullSubmissionID, value := range results {
		if value == nil {
			delete(results, fullSubmissionID)
		}
	}

	return results, err
}

func (this *backend) StoreIndividualAnalysis(records []*model.IndividualAnalysis) error {
	this.analysisIndividualLock.Lock()
	defer this.analysisIndividualLock.Unlock()

	return util.AppendJSONLFileMany(this.getAnalysisIndividualPath(), records)
}

func (this *backend) GetPairwiseAnalysis(keys []model.PairwiseKey) (map[model.PairwiseKey]*model.PairwiseAnalysis, error) {
	this.analysisPairwiseLock.RLock()
	defer this.analysisPairwiseLock.RUnlock()

	// Build a lookup that we will also use for storage.
	results := make(map[model.PairwiseKey]*model.PairwiseAnalysis, len(keys))
	for _, key := range keys {
		results[key] = nil
	}

	path := this.getAnalysisPairwisePath()

	err := util.ApplyJSONLFile(path, model.PairwiseAnalysis{}, func(index int, record *model.PairwiseAnalysis) {
		_, ok := results[record.SubmissionIDs]
		if ok {
			results[record.SubmissionIDs] = record
		}
	})

	// Cleanup missing keys.
	for key, value := range results {
		if value == nil {
			delete(results, key)
		}
	}

	return results, err
}

func (this *backend) StorePairwiseAnalysis(records []*model.PairwiseAnalysis) error {
	this.analysisPairwiseLock.Lock()
	defer this.analysisPairwiseLock.Unlock()

	return util.AppendJSONLFileMany(this.getAnalysisPairwisePath(), records)
}

func (this *backend) getAnalysisIndividualPath() string {
	return filepath.Join(this.baseDir, DISK_DB_ANALYSIS_INDIVIDUAL_FILENAME)
}

func (this *backend) getAnalysisPairwisePath() string {
	return filepath.Join(this.baseDir, DISK_DB_ANALYSIS_PAIRWISE_FILENAME)
}
