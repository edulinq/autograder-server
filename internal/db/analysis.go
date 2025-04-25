package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/model"
)

func GetIndividualAnalysis(fullSubmissionIDs []string) (map[string]*model.IndividualAnalysis, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetIndividualAnalysis(fullSubmissionIDs)
}

func GetPairwiseAnalysis(keys []model.PairwiseKey) (map[model.PairwiseKey]*model.PairwiseAnalysis, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetPairwiseAnalysis(keys)
}

// Get a single individual analysis result.
// If the id is not matched, return nil.
func GetSingleIndividualAnalysis(fullSubmissionID string) (*model.IndividualAnalysis, bool, error) {
	results, err := GetIndividualAnalysis([]string{fullSubmissionID})
	if err != nil {
		return nil, false, err
	}

	result, ok := results[fullSubmissionID]

	return result, ok, nil
}

// Get a single pairwise analysis result.
// If the key is not matched, return nil.
func GetSinglePairwiseAnalysis(key model.PairwiseKey) (*model.PairwiseAnalysis, bool, error) {
	results, err := GetPairwiseAnalysis([]model.PairwiseKey{key})
	if err != nil {
		return nil, false, err
	}

	result, ok := results[key]

	return result, ok, nil
}

func RemoveIndividualAnalysis(fullSubmissionIDs []string) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.RemoveIndividualAnalysis(fullSubmissionIDs)
}

func RemovePairwiseAnalysis(keys []model.PairwiseKey) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.RemovePairwiseAnalysis(keys)
}

func StoreIndividualAnalysis(records []*model.IndividualAnalysis) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StoreIndividualAnalysis(records)
}

func StorePairwiseAnalysis(records []*model.PairwiseAnalysis) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StorePairwiseAnalysis(records)
}
