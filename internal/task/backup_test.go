package task

import (
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model/tasks"
	"github.com/edulinq/autograder/internal/util"
)

const EXPECTED_MD5 = "b6a643542e26517e605eded2da6f8f91"

func TestBackupTempDir(test *testing.T) {
	tempDir, err := util.MkDirTemp("autograder-test-task-backup-")
	if err != nil {
		test.Fatalf("Failed to create temp dir: '%v'.", err)
	}

	doBackup(test, tempDir, filepath.Join(tempDir, "course101-test.zip"))
}

func TestBackupDefaultDir(test *testing.T) {
	doBackup(test, "", filepath.Join(config.GetBackupDir(), "course101-test.zip"))
}

func TestBackupOptionsDir(test *testing.T) {
	tempDir, err := util.MkDirTemp("autograder-test-task-backup-")
	if err != nil {
		test.Fatalf("Failed to create temp dir: '%v'.", err)
	}

	oldValue := config.TASK_BACKUP_DIR.Get()
	config.TASK_BACKUP_DIR.Set(tempDir)
	defer config.TASK_BACKUP_DIR.Set(oldValue)

	doBackup(test, "", filepath.Join(tempDir, "course101-test.zip"))
}

func doBackup(test *testing.T, dest string, expectedPath string) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	course := db.MustGetTestCourse()

	task := &tasks.BackupTask{
		BaseTask: &tasks.BaseTask{
			Disable: false,
			When:    []*common.ScheduledTime{},
		},
		Dest:     dest,
		BackupID: "test",
	}

	_, err := RunBackupTask(course, task)
	if err != nil {
		test.Fatalf("Failed to run backup task: '%v'.", err)
	}

	if !util.PathExists(expectedPath) {
		test.Fatalf("Could not find backup at expected location: '%s'.", expectedPath)
	}

	actualMD5, err := util.MD5FileHex(expectedPath)
	if err != nil {
		test.Fatalf("Failed to get MD5 from backup file: '%v'.", err)
	}

	if EXPECTED_MD5 != actualMD5 {
		test.Fatalf("MD5s do not match. Expected: '%s', Actual: '%s'.", EXPECTED_MD5, actualMD5)
	}
}
