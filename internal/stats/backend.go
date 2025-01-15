package stats

import (
	"sync"
)

var backend StorageBackend = nil
var backendLock sync.RWMutex

type StorageBackend interface {
	StoreSystemStats(record *SystemMetrics) error
	StoreCourseMetric(record *CourseMetric) error
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

func StoreCourseMetric(record *CourseMetric) error {
	backendLock.RLock()
	defer backendLock.RUnlock()

	if backend == nil {
		return nil
	}

	return backend.StoreCourseMetric(record)
}
