package disk

import (
	"path/filepath"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

const SYSTEM_STATS_FILENAME = "stats.jsonl"

func (this *backend) StoreSystemStats(record *stats.SystemMetrics) error {
	this.statsLock.Lock()
	defer this.statsLock.Unlock()

	return util.AppendJSONLFile(this.getSystemStatsPath(), record)
}

func (this *backend) GetSystemStats(query stats.Query) ([]*stats.SystemMetrics, error) {
	this.statsLock.RLock()
	defer this.statsLock.RUnlock()

	path := this.getSystemStatsPath()
	if !util.PathExists(path) {
		return make([]*stats.SystemMetrics, 0), nil
	}

	records, err := util.FilterJSONLFile(path, stats.SystemMetrics{}, func(record *stats.SystemMetrics) bool {
		return query.Match(record)
	})

	return records, err
}

func (this *backend) getSystemStatsPath() string {
	return filepath.Join(this.baseDir, SYSTEM_STATS_FILENAME)
}
