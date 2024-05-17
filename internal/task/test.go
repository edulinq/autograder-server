package task

import (
    "fmt"

    "github.com/edulinq/autograder/internal/model"
    "github.com/edulinq/autograder/internal/model/tasks"
)

func RunTestTask(course *model.Course, rawTask tasks.ScheduledTask) (bool, error) {
    task, ok := rawTask.(*tasks.TestTask);
    if (!ok) {
        return false, fmt.Errorf("Task is not a TestTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return true, nil;
    }

    return true, task.Func(task.Payload);
}
