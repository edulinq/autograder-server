package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

func GetLogRecords(query log.ParsedLogQuery) ([]*log.Record, error) {
	return GetLogRecordsFull(query, false)
}

func GetLogRecordsFull(query log.ParsedLogQuery, useTestingRecords bool) ([]*log.Record, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	if !useTestingRecords {
		return backend.GetLogRecords(query)
	}

	records := make([]*log.Record, 0)
	for _, record := range TESTING_LOG_RECORDS {
		if query.Match(record) {
			records = append(records, record)
		}
	}

	return records, nil
}

var TESTING_LOG_RECORDS []*log.Record = []*log.Record{
	&log.Record{
		Level:      log.LevelTrace,
		Message:    "Trace Course Log",
		Timestamp:  timestamp.Timestamp(100),
		Error:      "",
		Course:     "course101",
		Assignment: "hw0",
		User:       "course-other@test.edulinq.org",
	},
	&log.Record{
		Level:      log.LevelTrace,
		Message:    "Trace Server Log",
		Timestamp:  timestamp.Timestamp(150),
		Error:      "",
		Course:     "",
		Assignment: "",
		User:       "server-user@test.edulinq.org",
	},
	&log.Record{
		Level:      log.LevelDebug,
		Message:    "Debug Course Log",
		Timestamp:  timestamp.Timestamp(200),
		Error:      "",
		Course:     "course101",
		Assignment: "hw0",
		User:       "course-student@test.edulinq.org",
	},
	&log.Record{
		Level:      log.LevelDebug,
		Message:    "Debug Server Log",
		Timestamp:  timestamp.Timestamp(250),
		Error:      "",
		Course:     "",
		Assignment: "",
		User:       "server-creator@test.edulinq.org",
	},
	&log.Record{
		Level:      log.LevelInfo,
		Message:    "Info Course Log",
		Timestamp:  timestamp.Timestamp(300),
		Error:      "",
		Course:     "course101",
		Assignment: "hw0",
		User:       "course-grader@test.edulinq.org",
	},
	&log.Record{
		Level:      log.LevelInfo,
		Message:    "Info Server Log",
		Timestamp:  timestamp.Timestamp(350),
		Error:      "",
		Course:     "",
		Assignment: "",
		User:       "server-admin@test.edulinq.org",
	},
	&log.Record{
		Level:      log.LevelWarn,
		Message:    "Warn Course Log",
		Timestamp:  timestamp.Timestamp(400),
		Error:      "Course Warning",
		Course:     "course101",
		Assignment: "hw0",
		User:       "course-admin@test.edulinq.org",
	},
	&log.Record{
		Level:      log.LevelWarn,
		Message:    "Warn Server Log",
		Timestamp:  timestamp.Timestamp(450),
		Error:      "Server Warning",
		Course:     "",
		Assignment: "",
		User:       "server-owner@test.edulinq.org",
	},
	&log.Record{
		Level:      log.LevelError,
		Message:    "Error Course Log",
		Timestamp:  timestamp.Timestamp(500),
		Error:      "Course Error",
		Course:     "course101",
		Assignment: "",
		User:       "course-owner@test.edulinq.org",
	},
	&log.Record{
		Level:      log.LevelError,
		Message:    "Error Server Log",
		Timestamp:  timestamp.Timestamp(550),
		Error:      "Server Error",
		Course:     "",
		Assignment: "",
		User:       "",
	},
	&log.Record{
		Level:      log.LevelFatal,
		Message:    "Fatal Course Log",
		Timestamp:  timestamp.Timestamp(600),
		Error:      "",
		Course:     "course101",
		Assignment: "",
		User:       "",
	},
	&log.Record{
		Level:      log.LevelFatal,
		Message:    "Fatal Server Log",
		Timestamp:  timestamp.Timestamp(650),
		Error:      "",
		Course:     "",
		Assignment: "",
		User:       "",
	},
}
