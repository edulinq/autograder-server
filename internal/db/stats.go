package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/stats"
)

func GetMetrics(query stats.Query) ([]*stats.Metric, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetMetrics(query)
}

func StoreMetric(record *stats.Metric) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StoreMetric(record)
}
