package disk

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const LOG_FILENAME = "log.jsonl"

func (this *backend) LogDirect(record *log.Record) error {
	this.logLock.Lock()
	defer this.logLock.Unlock()

	line, err := util.ToJSON(record)
	if err != nil {
		return fmt.Errorf("Failed to convert log record to JSON: '%w'.", err)
	}

	path := this.getLogPath()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to open log file '%s': '%w'.", path, err)
	}
	defer file.Close()

	_, err = file.WriteString(line + "\n")
	if err != nil {
		return fmt.Errorf("Failed to write record to log file '%s': '%w'.", path, err)
	}

	return nil
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
