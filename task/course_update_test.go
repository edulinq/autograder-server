package task

import (
    "testing"

    "github.com/edulinq/autograder/common"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/model/tasks"
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

    _, err := RunCourseUpdateTask(course, task);
    if (err != nil) {
        test.Fatalf("Failed to run course update task: '%v'.", err);
    }
}
