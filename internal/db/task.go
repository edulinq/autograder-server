package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/timestamp"
)

func LogTaskCompletion(courseID string, taskID string, instance timestamp.Timestamp) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.LogTaskCompletion(courseID, taskID, instance)
}

func GetLastTaskCompletion(courseID string, taskID string) (timestamp.Timestamp, error) {
	if backend == nil {
		return timestamp.Zero(), fmt.Errorf("Database has not been opened.")
	}

	return backend.GetLastTaskCompletion(courseID, taskID)
}
