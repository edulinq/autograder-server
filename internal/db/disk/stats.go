package disk

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

const SYSTEM_STATS_FILENAME = "stats.jsonl"
const COURSE_STATS_FILENAME = "course-stats.jsonl"
const API_REQUEST_STATS_FILENAME = "api-request-stats.jsonl"

func (this *backend) GetSystemStats(query stats.Query) ([]*stats.SystemMetrics, error) {
	path := this.getSystemStatsPath()

	this.statsLock.RLock()
	defer this.statsLock.RUnlock()

	records, err := util.FilterJSONLFile(path, stats.SystemMetrics{}, func(record *stats.SystemMetrics) bool {
		return query.Match(record)
	})

	return records, err
}

func (this *backend) StoreSystemStats(record *stats.SystemMetrics) error {
	this.statsLock.Lock()
	defer this.statsLock.Unlock()

	return util.AppendJSONLFile(this.getSystemStatsPath(), record)
}

func (this *backend) GetCourseMetrics(query stats.MetricQuery) ([]*stats.BaseMetric, error) {
	courseID, ok := query.Where[stats.COURSE_ID_KEY].(string)
	if !ok {
		return nil, fmt.Errorf("When querying for course metrics, course ID must not be empty.")
	}

	path := this.getCourseStatsPath(courseID)

	this.contextReadLock(path)
	defer this.contextReadUnlock(path)

	records, err := util.FilterJSONLFile(path, stats.BaseMetric{}, func(record *stats.BaseMetric) bool {
		return query.Match(record)
	})

	return records, err
}

func (this *backend) StoreCourseMetric(record *stats.BaseMetric) error {
	courseID, ok := record.Attributes[stats.COURSE_ID_KEY].(string)
	if !ok {
		return fmt.Errorf("No course ID was given.")
	}

	path := this.getCourseStatsPath(courseID)

	this.contextLock(path)
	defer this.contextUnlock(path)

	return util.AppendJSONLFile(path, record)
}

func (this *backend) GetAPIRequestMetrics(query stats.MetricQuery) ([]*stats.BaseMetric, error) {
	path := this.getAPIRequestStatsPath()

	this.apiRequestLock.RLock()
	defer this.apiRequestLock.RUnlock()

	records, err := util.FilterJSONLFile(path, stats.BaseMetric{}, func(record *stats.BaseMetric) bool {
		return query.Match(record)
	})

	return records, err
}

func (this *backend) StoreAPIRequestMetric(record *stats.BaseMetric) error {
	path := this.getAPIRequestStatsPath()

	this.apiRequestLock.Lock()
	defer this.apiRequestLock.Unlock()

	return util.AppendJSONLFile(path, record)
}

func (this *backend) getSystemStatsPath() string {
	return filepath.Join(this.baseDir, SYSTEM_STATS_FILENAME)
}

func (this *backend) getCourseStatsPath(courseID string) string {
	return filepath.Join(this.getCourseDirFromID(courseID), COURSE_STATS_FILENAME)
}

func (this *backend) getAPIRequestStatsPath() string {
	return filepath.Join(this.baseDir, API_REQUEST_STATS_FILENAME)
}
