package task

import (
    "fmt"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/model/tasks"
    "github.com/eriq-augustine/autograder/scoring"
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
