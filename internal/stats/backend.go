package stats

import (
	"sync"
)

var backend StorageBackend = nil
var backendLock sync.RWMutex

type StorageBackend interface {
	GetSystemStats(query Query) ([]*SystemMetrics, error)
	StoreSystemStats(record *SystemMetrics) error

	GetMetrics(query Query) ([]*Metric, error)
	StoreMetric(record *Metric) error
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

func GetMetric(query Query) ([]*Metric, error) {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil, nil
	}

	return backend.GetMetrics(query)
}

func StoreMetric(record *Metric) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreMetric(record)
}
