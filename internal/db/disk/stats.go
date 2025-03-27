package disk

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

const SYSTEM_STATS_FILENAME = "stats.jsonl"

func (this *backend) GetSystemStats(query stats.Query) ([]*stats.SystemMetrics, error) {
	path := this.getSystemStatsPath()

	this.systemStatsLock.RLock()
	defer this.systemStatsLock.RUnlock()

	records, err := util.FilterJSONLFile(path, stats.SystemMetrics{}, func(record *stats.SystemMetrics) bool {
		return query.Match(record)
	})

	return records, err
}

func (this *backend) StoreSystemStats(record *stats.SystemMetrics) error {
	this.systemStatsLock.Lock()
	defer this.systemStatsLock.Unlock()

	return util.AppendJSONLFile(this.getSystemStatsPath(), record)
}

func (this *backend) GetMetrics(query stats.MetricQuery) ([]*stats.BaseMetric, error) {
	path, err := this.getStatsPath(query.Type)
	if err != nil {
		return nil, err
	}

	this.statsLock.Lock()
	defer this.statsLock.Unlock()

	records, err := util.FilterJSONLFile(path, stats.BaseMetric{}, func(record *stats.BaseMetric) bool {
		return query.Match(record)
	})

	return records, err
}

func (this *backend) StoreMetric(record *stats.BaseMetric) error {
	path, err := this.getStatsPath(record.Type)
	if err != nil {
		return err
	}

	// if (record.Type == stats.GRADING_TIME_STATS_KEY || record.Type == stats.TASK_TIME_STATS_KEY){
	// 	course, ok := record.Attributes[stats.COURSE_ID_KEY]
	// 	if !ok {
	// 		return fmt.Errorf("")
	// 	}
	// }

	this.statsLock.Lock()
	defer this.statsLock.Unlock()

	return util.AppendJSONLFile(path, record)
}

func (this *backend) getStatsPath(metricType string) (string, error) {
	if metricType == "" {
		return "", fmt.Errorf("No metric type was given.")
	}

	return filepath.Join(this.baseDir, metricType+".jsonl"), nil
}

func (this *backend) getSystemStatsPath() string {
	return filepath.Join(this.baseDir, SYSTEM_STATS_FILENAME)
}
