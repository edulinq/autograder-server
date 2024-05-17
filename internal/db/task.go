package db

import (
    "fmt"
    "time"
)

func LogTaskCompletion(courseID string, taskID string, instance time.Time) error {
    if (backend == nil) {
        return fmt.Errorf("Database has not been opened.");
    }

    return backend.LogTaskCompletion(courseID, taskID, instance);
}

func GetLastTaskCompletion(courseID string, taskID string) (time.Time, error) {
    if (backend == nil) {
        return time.Time{}, fmt.Errorf("Database has not been opened.");
    }

    return backend.GetLastTaskCompletion(courseID, taskID);
}
