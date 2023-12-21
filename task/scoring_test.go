package task

import (
    "testing"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model/tasks"
)

func TestScoringUploadBase(test *testing.T) {
    db.ResetForTesting();

    course := db.MustGetTestCourse();

    task := &tasks.ScoringUploadTask{
        BaseTask: &tasks.BaseTask{
            Disable: false,
            When: []*common.ScheduledTime{},
        },
        DryRun: true,
    };

    _, err := RunScoringUploadTask(course, task);
    if (err != nil) {
        test.Fatalf("Failed to run scoring upload task: '%v'.", err);
    }
}
