package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/model"
)

// Get a single pairwise analysis result.
// If the key is not matched, return nil.
func GetSinglePairwiseAnalysis(key model.PairwiseKey) (*model.PairwiseAnalysis, error) {
	results, err := GetPairwiseAnalysis([]model.PairwiseKey{key})
	if err != nil {
		return nil, err
	}

	return results[key], nil
}

func GetPairwiseAnalysis(keys []model.PairwiseKey) (map[model.PairwiseKey]*model.PairwiseAnalysis, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetPairwiseAnalysis(keys)
}

func StorePairwiseAnalysis(records []*model.PairwiseAnalysis) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StorePairwiseAnalysis(records)
}
