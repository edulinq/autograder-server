package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/model"
)

func GetActiveCourseTasks(course *model.Course) (map[string]*model.FullScheduledTask, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetActiveCourseTasks(course)
}

func GetActiveTasks() (map[string]*model.FullScheduledTask, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetActiveTasks()
}

func GetNextActiveTask() (*model.FullScheduledTask, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetNextActiveTask()
}

// Upsert an active task.
// This can be used to record that a task has been completed (upsert with updated timestamps).
func UpsertActiveTask(task *model.FullScheduledTask) error {
	tasks := map[string]*model.FullScheduledTask{
		task.Hash: task,
	}

	return UpsertActiveTasks(tasks)
}

func UpsertActiveTasks(tasks map[string]*model.FullScheduledTask) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.UpsertActiveTasks(tasks)
}
