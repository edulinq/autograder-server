package disk

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

const SYSTEM_STATS_FILENAME = "system-stats.jsonl"

func (this *backend) GetSystemStats(query stats.Query) ([]*stats.SystemMetrics, error) {
	path := this.getSystemStatsPath()

	this.systemStatsLock.RLock()
	defer this.systemStatsLock.RUnlock()

	records, err := util.FilterJSONLFile(path, stats.SystemMetrics{}, func(record *stats.SystemMetrics) bool {
		return query.MatchTimeWindow(record)
	})

	return records, err
}

func (this *backend) StoreSystemStats(record *stats.SystemMetrics) error {
	this.systemStatsLock.Lock()
	defer this.systemStatsLock.Unlock()

	return util.AppendJSONLFile(this.getSystemStatsPath(), record)
}

func (this *backend) GetMetrics(query stats.Query) ([]*stats.Metric, error) {
	path, err := this.getStatsPath(query.Type)
	if err != nil {
		return nil, err
	}

	this.contextReadLock(path)
	defer this.contextReadUnlock(path)

	records, err := util.FilterJSONLFile(path, stats.Metric{}, func(record *stats.Metric) bool {
		return query.Match(record)
	})

	return records, err
}

func (this *backend) StoreMetric(record *stats.Metric) error {
	path, err := this.getStatsPath(record.Type)
	if err != nil {
		return err
	}

	this.contextLock(path)
	defer this.contextUnlock(path)

	return util.AppendJSONLFile(path, record)
}

func (this *backend) getStatsPath(metricType stats.MetricType) (string, error) {
	if metricType == "" {
		return "", fmt.Errorf("No metric type was given.")
	}

	path := filepath.Join(this.baseDir, string(metricType)+".jsonl")
	path = util.ShouldNormalizePath(path)

	return path, nil
}

func (this *backend) getSystemStatsPath() string {
	return filepath.Join(this.baseDir, SYSTEM_STATS_FILENAME)
}
