package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/stats"
)

func GetSystemStats(query stats.Query) ([]*stats.SystemMetrics, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetSystemStats(query)
}

func StoreSystemStats(record *stats.SystemMetrics) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StoreSystemStats(record)
}

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
