package task

import (
    "testing"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model/tasks"
)

func TestCourseUpdateTaskBase(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    course := db.MustGetTestCourse();

    task := &tasks.CourseUpdateTask{
        BaseTask: &tasks.BaseTask{
            Disable: false,
            When: []*common.ScheduledTime{},
        },
    };

    err := RunCourseUpdateTask(course, task);
    if (err != nil) {
        test.Fatalf("Failed to run course update task: '%v'.", err);
    }
}
