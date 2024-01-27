package disk

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"

    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/util"
)

const LOG_FILENAME = "log.jsonl"

func (this *backend) LogDirect(record *log.Record) error {
    this.lock.Lock();
    defer this.lock.Unlock();

    this.logLock.Lock();
    defer this.logLock.Unlock();

    line, err := util.ToJSON(record);
    if (err != nil) {
        return fmt.Errorf("Failed to convert log record to JSON: '%w'.", err);
    }

    path := this.getLogPath();
	file, err := os.OpenFile(path, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	if (err != nil) {
        return fmt.Errorf("Failed to open log file '%s': '%w'.", path, err);
	}
    defer file.Close();

	_, err = file.WriteString(line + "\n");
    if (err != nil) {
        return fmt.Errorf("Failed to write record to log file '%s': '%w'.", path, err);
	}

    return nil;
}

func (this *backend) GetLogRecords(level log.LogLevel, after time.Time, courseID string, assignmentID string, userID string) ([]*log.Record, error) {
    this.logLock.RLock();
    defer this.logLock.RUnlock();

    records := make([]*log.Record, 0);

    path := this.getLogPath();
    if (!util.PathExists(path)) {
        return records, nil;
    }

    file, err := os.Open(path);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to open log file '%s': '%w'.", path, err);
    }
    defer file.Close();

    lineno := 0;
    reader := bufio.NewReader(file);
    for {
        line, err := readline(reader);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to read line from log file '%s': '%w'.", path, err);
        }

        if (line == nil) {
            // EOF.
            break;
        }

        lineno++;

        var record log.Record;
        err = util.JSONFromBytes(line, &record);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to convert log line %d from file '%s' to JSON: '%w'.", lineno, path, err);
        }

        keep, err := keepRecord(&record, level, after, courseID, assignmentID, userID);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to filter log line %d from file '%s': '%w'.", lineno, path, err);
        }

        if (!keep) {
            continue;
        }

        records = append(records, &record);
    }

    return records, nil;
}

func keepRecord(record *log.Record, level log.LogLevel, after time.Time, courseID string, assignmentID string, userID string) (bool, error) {
    if (record.Level < level) {
        return false, nil;
    }

    if ((courseID != "") && (record.Course != courseID)) {
        return false, nil;
    }

    if ((assignmentID != "") && (record.Assignment != assignmentID)) {
        return false, nil;
    }

    if ((userID != "") && (record.User != userID)) {
        return false, nil;
    }

    if (!after.IsZero()) {
        recordTime := time.UnixMicro(record.UnixMicro);
        if (!recordTime.After(after)) {
            return false, nil;
        }
    }

    return true, nil;
}

// Will only return a nil content or error or EOF.
func readline(reader *bufio.Reader) ([]byte, error) {
    var isPrefix bool = true;
    var err error;

    var line []byte;
    var fullLine []byte;

    for (isPrefix && err == nil) {
        line, isPrefix, err = reader.ReadLine();
        fullLine = append(fullLine, line...);
    }

    if (err == io.EOF) {
        if (fullLine == nil) {
            return nil, nil;
        }

        return fullLine, nil;
    }

    return fullLine, err;
}

func (this *backend) getLogPath() string {
    return filepath.Join(this.baseDir, LOG_FILENAME);
}
