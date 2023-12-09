package task

import (
    "testing"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model/tasks"
)

func TestScoringUploadBase(test *testing.T) {
    db.ResetForTesting();

    course := db.MustGetTestCourse();

    task := &tasks.ScoringUploadTask{
        BaseTask: &tasks.BaseTask{
            Disable: false,
            When: []*tasks.ScheduledTime{},
        },
        DryRun: true,
    };

    err := RunScoringUploadTask(course, task);
    if (err != nil) {
        test.Fatalf("Failed to run scoring upload task: '%v'.", err);
    }
}
