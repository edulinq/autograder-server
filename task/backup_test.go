package task

import (
    "path/filepath"
    "testing"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model/tasks"
    "github.com/eriq-augustine/autograder/util"
)

func TestBackupBase(test *testing.T) {
    db.ResetForTesting();

    expectedMD5 := "e5169a963b347c148b7ec6bf8657a305";

    course := db.MustGetTestCourse();

    tempDir, err := util.MkDirTemp("autograder-test-task-backup-");
    if (err != nil) {
        test.Fatalf("Failed to create temp dir: '%v'.", err);
    }
    defer util.RemoveDirent(tempDir)

    task := &tasks.BackupTask{
        BaseTask: &tasks.BaseTask{
            Disable: false,
            When: []*tasks.ScheduledTime{},
        },
        Dest: tempDir,
        BackupID: "test",
    };

    err = RunBackupTask(course, task);
    if (err != nil) {
        test.Fatalf("Failed to run backup task: '%v'.", err);
    }

    path := filepath.Join(tempDir, "course101-test.zip");
    actualMD5, err := util.MD5FileHex(path);
    if (err != nil) {
        test.Fatalf("Failed to get MD5 from backup file: '%v'.", err);
    }

    if (expectedMD5 != actualMD5) {
        test.Fatalf("MD5s do not match. Expected: '%s', Actual: '%s'.", expectedMD5, actualMD5);
    }
}
