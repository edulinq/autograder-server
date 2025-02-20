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

func GetCourseMetrics(query stats.CourseMetricQuery) ([]*stats.CourseMetric, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetCourseMetrics(query)
}

func GetRequestMetrics(query stats.Query) ([]*stats.RequestMetric, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetRequestMetrics(query)
}

func StoreSystemStats(record *stats.SystemMetrics) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StoreSystemStats(record)
}

func StoreCourseMetric(record *stats.CourseMetric) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StoreCourseMetric(record)
}

func StoreRequestMetric(record *stats.RequestMetric) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StoreRequestMetric(record)
}
