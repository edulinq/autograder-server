package disk

import (
	"path/filepath"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const LOG_FILENAME = "log.jsonl"

func (this *backend) LogDirect(record *log.Record) error {
	this.logLock.Lock()
	defer this.logLock.Unlock()

	return util.AppendJSONLFile(this.getLogPath(), record)
}

func (this *backend) GetLogRecords(query log.ParsedLogQuery) ([]*log.Record, error) {
	this.logLock.RLock()
	defer this.logLock.RUnlock()

	path := this.getLogPath()
	if !util.PathExists(path) {
		return make([]*log.Record, 0), nil
	}

	records, err := util.FilterJSONLFile(path, log.Record{}, func(record *log.Record) bool {
		return keepRecord(record, query)
	})

	return records, err
}

func keepRecord(record *log.Record, query log.ParsedLogQuery) bool {
	if record.Level < query.Level {
		return false
	}

	if (query.CourseID != "") && (record.Course != query.CourseID) {
		return false
	}

	// Assignment ID will only be matched on if the course ID also matches.
	courseMatch := ((query.CourseID != "") && (record.Course == query.CourseID))

	if (query.AssignmentID != "") && (!courseMatch || (record.Assignment != query.AssignmentID)) {
		return false
	}

	if (query.UserEmail != "") && (record.User != query.UserEmail) {
		return false
	}

	if record.Timestamp < query.After {
		return false
	}

	return true
}

func (this *backend) getLogPath() string {
	return filepath.Join(this.baseDir, LOG_FILENAME)
}
