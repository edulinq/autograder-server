package task

import (
    "fmt"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/model/tasks"
)

func RunCourseUpdateTask(course *model.Course, rawTask tasks.ScheduledTask) error {
    task, ok := rawTask.(*tasks.CourseUpdateTask);
    if (!ok) {
        return fmt.Errorf("Task is not a CourseUpdateTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return nil;
    }

    _, _, err := db.UpdateCourseFromSource(course);
    return err;
}
