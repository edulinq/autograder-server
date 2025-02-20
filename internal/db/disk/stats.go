package disk

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

const SYSTEM_STATS_FILENAME = "stats.jsonl"
const COURSE_STATS_FILENAME = "course-stats.jsonl"
const REQUEST_STATS_FILENAME = "request-stats.jsonl"

func (this *backend) StoreSystemStats(record *stats.SystemMetrics) error {
	this.statsLock.Lock()
	defer this.statsLock.Unlock()

	return util.AppendJSONLFile(this.getSystemStatsPath(), record)
}

func (this *backend) GetSystemStats(query stats.Query) ([]*stats.SystemMetrics, error) {
	this.statsLock.RLock()
	defer this.statsLock.RUnlock()

	path := this.getSystemStatsPath()

	records, err := util.FilterJSONLFile(path, stats.SystemMetrics{}, func(record *stats.SystemMetrics) bool {
		return query.Match(record)
	})

	return records, err
}

func (this *backend) StoreCourseMetric(record *stats.CourseMetric) error {
	path := this.getCourseStatsPath(record.CourseID)

	this.contextLock(path)
	defer this.contextUnlock(path)

	return util.AppendJSONLFile(path, record)
}

func (this *backend) GetCourseMetrics(query stats.CourseMetricQuery) ([]*stats.CourseMetric, error) {
	if query.CourseID == "" {
		return nil, fmt.Errorf("When querying for course metrics, course ID must not be empty.")
	}

	path := this.getCourseStatsPath(query.CourseID)

	this.contextReadLock(path)
	defer this.contextReadUnlock(path)

	records, err := util.FilterJSONLFile(path, stats.CourseMetric{}, func(record *stats.CourseMetric) bool {
		return query.Match(record)
	})

	return records, err
}

func (this *backend) StoreRequestMetric(record *stats.RequestMetric) error {
	path := this.getRequestStatsPath()

	this.requestLock.Lock()
	defer this.requestLock.Unlock()

	return util.AppendJSONLFile(path, record)
}

func (this *backend) getSystemStatsPath() string {
	return filepath.Join(this.baseDir, SYSTEM_STATS_FILENAME)
}

func (this *backend) getCourseStatsPath(courseID string) string {
	return filepath.Join(this.getCourseDirFromID(courseID), COURSE_STATS_FILENAME)
}

func (this *backend) getRequestStatsPath() string {
	return filepath.Join(this.baseDir, REQUEST_STATS_FILENAME)
}
