package task

import (
    "fmt"

    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/model/tasks"
    "github.com/edulinq/autograder/scoring"
)

func RunScoringUploadTask(course *model.Course, rawTask tasks.ScheduledTask) (bool, error) {
    task, ok := rawTask.(*tasks.ScoringUploadTask);
    if (!ok) {
        return false, fmt.Errorf("Task is not a ScoringUploadTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return true, nil;
    }

    return true, scoring.FullCourseScoringAndUpload(course, task.DryRun);
}
