package disk

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

const SYSTEM_STATS_FILENAME = "stats.jsonl"

func (this *backend) StoreSystemStats(record *stats.SystemMetrics) error {
	this.statsLock.Lock()
	defer this.statsLock.Unlock()

	line, err := util.ToJSON(record)
	if err != nil {
		return fmt.Errorf("Failed to convert stats record to JSON: '%w'.", err)
	}

	path := this.getSystemStatsPath()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to open stats file '%s': '%w'.", path, err)
	}
	defer file.Close()

	_, err = file.WriteString(line + "\n")
	if err != nil {
		return fmt.Errorf("Failed to write record to stats file '%s': '%w'.", path, err)
	}

	return nil
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
