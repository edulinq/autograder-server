package task

import (
    "fmt"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/scoring"
)

func RunScoringUploadTask(course model.Course, rawTask model.ScheduledTask) error {
    task, ok := rawTask.(*model.ScoringUploadTask);
    if (!ok) {
        return fmt.Errorf("Task is not a ScoringUploadTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return nil;
    }

    return scoring.FullCourseScoringAndUpload(course, task.DryRun);
}
