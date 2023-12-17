package task

import (
    "fmt"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/model/tasks"
)

func RunTestTask(course *model.Course, rawTask tasks.ScheduledTask) error {
    task, ok := rawTask.(*tasks.TestTask);
    if (!ok) {
        return fmt.Errorf("Task is not a TestTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return nil;
    }

    return task.Func(task.Payload);
}
