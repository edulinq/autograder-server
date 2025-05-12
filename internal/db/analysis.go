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
