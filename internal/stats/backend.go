package stats

var backend StorageBackend = nil

func Init(newBackend StorageBackend, systemIntervalMS int) {
	backend = newBackend

	startSystemStatsCollection(systemIntervalMS)
}

type StorageBackend interface {
	StoreSystemMetrics(record *SystemMetrics) error
}
