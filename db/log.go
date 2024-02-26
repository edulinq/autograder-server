package db

import (
	"fmt"
	"time"

	"github.com/edulinq/autograder/log"
)

func GetLogRecords(level log.LogLevel, after time.Time, courseID string, assignmentID string, userID string) ([]*log.Record, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetLogRecords(level, after, courseID, assignmentID, userID)
}
