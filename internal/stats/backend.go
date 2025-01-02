package stats

var backend StorageBackend = nil

type StorageBackend interface {
	StoreSystemStats(record *SystemMetrics) error
}

func SetStorageBackend(newBackend StorageBackend) {
	backend = newBackend
}

func StartCollection(systemIntervalMS int) {
	startSystemStatsCollection(systemIntervalMS)
}

func StopCollection() {
	stopSystemStatsCollection()
}
