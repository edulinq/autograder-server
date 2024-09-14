package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/log"
)

func GetLogRecords(query log.ParsedLogQuery) ([]*log.Record, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetLogRecords(query)
}
