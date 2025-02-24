package stats

import (
	"sync"
)

var backend StorageBackend = nil
var backendLock sync.RWMutex

type StorageBackend interface {
	StoreSystemStats(record *SystemMetrics) error
	GetSystemStats(query Query) ([]*SystemMetrics, error)
	StoreCourseMetric(record *CourseMetric) error
	GetCourseMetrics(query CourseMetricQuery) ([]*CourseMetric, error)
	StoreAPIRequestMetric(record *APIRequestMetric) error
	GetAPIRequestMetrics(query Query) ([]*APIRequestMetric, error)
}

func SetStorageBackend(newBackend StorageBackend) {
	backendLock.Lock()
	defer backendLock.Unlock()

	backend = newBackend
}

func StartCollection(systemIntervalMS int) {
	startSystemStatsCollection(systemIntervalMS)
}

func StopCollection() {
	stopSystemStatsCollection()
}

func storeSystemStats(record *SystemMetrics) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreSystemStats(record)
}

func GetSystemStats(query Query) ([]*SystemMetrics, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetSystemStats(query)
}

func StoreCourseMetric(record *CourseMetric) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreCourseMetric(record)
}

func GetCourseMetrics(query CourseMetricQuery) ([]*CourseMetric, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetCourseMetrics(query)
}

func StoreAPIRequestMetric(record *APIRequestMetric) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreAPIRequestMetric(record)
}

func GetAPIRequestMetrics(query Query) ([]*APIRequestMetric, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetAPIRequestMetrics(query)
}
