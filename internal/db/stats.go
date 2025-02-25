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

func GetCourseMetrics(query stats.CourseMetricQuery) ([]*stats.CourseMetric, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetCourseMetrics(query)
}

func StoreCourseMetric(record *stats.CourseMetric) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StoreCourseMetric(record)
}

func GetAPIRequestMetrics(query stats.APIRequestMetricQuery) ([]*stats.APIRequestMetric, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetAPIRequestMetrics(query)
}

func StoreAPIRequestMetric(record *stats.APIRequestMetric) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StoreAPIRequestMetric(record)
}
