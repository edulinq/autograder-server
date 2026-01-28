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
		return query.Match(record)
	})

	return records, err
}

func (this *backend) getLogPath() string {
	return filepath.Join(this.baseDir, LOG_FILENAME)
}
