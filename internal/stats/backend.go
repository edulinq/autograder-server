package stats

import (
	"sync"
)

var backend StorageBackend = nil
var backendLock sync.RWMutex

type StorageBackend interface {
	GetSystemStats(query Query) ([]*SystemMetrics, error)
	StoreSystemStats(record *SystemMetrics) error

	GetCourseMetrics(query MetricQuery) ([]*BaseMetric, error)
	StoreCourseMetric(record *BaseMetric) error

	GetAPIRequestMetrics(query MetricQuery) ([]*BaseMetric, error)
	StoreAPIRequestMetric(record *BaseMetric) error
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

func GetCourseMetrics(query MetricQuery) ([]*BaseMetric, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetCourseMetrics(query)
}

func StoreCourseMetric(record *BaseMetric) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreCourseMetric(record)
}

func GetAPIRequestMetrics(query MetricQuery) ([]*BaseMetric, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetAPIRequestMetrics(query)
}

func StoreAPIRequestMetric(record *BaseMetric) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreAPIRequestMetric(record)
}
