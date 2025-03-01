package backup

import (
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

// This hash is expected to change when the test data for course101 is changed.
const EXPECTED_MD5 = "cb6ae9f8ed95ac4ea2e200a673659078"

func TestBackupTempDir(test *testing.T) {
	tempDir, err := util.MkDirTemp("autograder-test-course-backup-")
	if err != nil {
		test.Fatalf("Failed to create temp dir: '%v'.", err)
	}
	defer util.RemoveDirent(tempDir)

	doBackup(test, tempDir, filepath.Join(tempDir, "course101-test.zip"))
}

func TestBackupDefaultDir(test *testing.T) {
	doBackup(test, "", filepath.Join(config.GetBackupDir(), "course101-test.zip"))
}

func TestBackupOptionsDir(test *testing.T) {
	tempDir, err := util.MkDirTemp("autograder-test-course-backup-")
	if err != nil {
		test.Fatalf("Failed to create temp dir: '%v'.", err)
	}
	defer util.RemoveDirent(tempDir)

	oldValue := config.BACKUP_DIR.Get()
	config.BACKUP_DIR.Set(tempDir)
	defer config.BACKUP_DIR.Set(oldValue)

	doBackup(test, "", filepath.Join(tempDir, "course101-test.zip"))
}

func doBackup(test *testing.T, dest string, expectedPath string) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	course := db.MustGetTestCourse()

	err := BackupCourseFull(course, dest, "test")
	if err != nil {
		test.Fatalf("Failed to run course backup: '%v'.", err)
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
