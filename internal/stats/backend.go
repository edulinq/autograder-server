package stats

import (
	"sync"
)

var backend StorageBackend = nil
var backendLock sync.RWMutex

type StorageBackend interface {
	GetSystemStats(query Query) ([]*SystemMetrics, error)
	StoreSystemStats(record *SystemMetrics) error

	GetCourseMetrics(query MetricQuery) ([]*CourseMetric, error)
	StoreCourseMetric(record *CourseMetric) error

	GetAPIRequestMetrics(query MetricQuery) ([]*APIRequestMetric, error)
	StoreAPIRequestMetric(record *APIRequestMetric) error
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

func GetSystemStats(query Query) ([]*SystemMetrics, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetSystemStats(query)
}

func storeSystemStats(record *SystemMetrics) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreSystemStats(record)
}

func GetCourseMetrics(query MetricQuery) ([]*CourseMetric, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetCourseMetrics(query)
}

func StoreCourseMetric(record *CourseMetric) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreCourseMetric(record)
}

func GetAPIRequestMetrics(query MetricQuery) ([]*APIRequestMetric, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetAPIRequestMetrics(query)
}

func StoreAPIRequestMetric(record *APIRequestMetric) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreAPIRequestMetric(record)
}
