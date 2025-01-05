package stats

import (
	"sync"
)

var backend StorageBackend = nil
var backendLock sync.Mutex

type StorageBackend interface {
	StoreSystemStats(record *SystemMetrics) error
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
	backendLock.Lock()
	defer backendLock.Unlock()

	if backend == nil {
		return nil
	}

	return backend.StoreSystemStats(record)
}
