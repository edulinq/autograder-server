package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

func GetLogRecords(level log.LogLevel, after timestamp.Timestamp, courseID string, assignmentID string, userID string) ([]*log.Record, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetLogRecords(level, after, courseID, assignmentID, userID)
}
