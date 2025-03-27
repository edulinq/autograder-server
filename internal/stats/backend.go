package stats

import (
	"sync"
)

var backend StorageBackend = nil
var backendLock sync.RWMutex

type StorageBackend interface {
	GetSystemStats(query Query) ([]*SystemMetrics, error)
	StoreSystemStats(record *SystemMetrics) error

	GetMetrics(query MetricQuery) ([]*BaseMetric, error)
	StoreMetric(record *BaseMetric) error
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

func GetMetric(query MetricQuery) ([]*BaseMetric, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetMetrics(query)
}

func StoreMetric(record *BaseMetric) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreMetric(record)
}
