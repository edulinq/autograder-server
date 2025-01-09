package task

import (
	"fmt"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/model/tasks"
	"github.com/edulinq/autograder/internal/scoring"
)

func RunScoringUploadTask(course *model.Course, rawTask tasks.ScheduledTask) (bool, error) {
	task, ok := rawTask.(*tasks.ScoringUploadTask)
	if !ok {
		return false, fmt.Errorf("Task is not a ScoringUploadTask: %t (%v).", rawTask, rawTask)
	}

	if task.Disable {
		return true, nil
	}

	_, err := scoring.FullCourseScoringAndUpload(course, task.DryRun)
	return true, err
}
