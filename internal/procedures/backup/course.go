package backup

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func BackupCourse(courseID string) error {
	course, err := db.GetCourse(courseID)
	if err != nil {
		return fmt.Errorf("Failed to get course ('%s') for backup: '%w'.", courseID, err)
	}

	if course == nil {
		return fmt.Errorf("Unable to find course ('%s') for backup.", courseID)
	}

	return BackupCourseFull(course, "", "")
}

func BackupCourseFull(course *model.Course, dest string, backupID string) error {
	if dest == "" {
		dest = config.GetBackupDir()
	}

	if util.IsFile(dest) {
		return fmt.Errorf("Backup directory exists and is a file: '%s'.", dest)
	}

	err := util.MkDir(dest)
	if err != nil {
		return fmt.Errorf("Could not create dest dir '%s': '%w'.", dest, err)
	}

	baseTempDir, err := util.MkDirTemp("autograder-backup-course-")
	if err != nil {
		return fmt.Errorf("Could not create temp backup dir: '%w'.", err)
	}
	defer util.RemoveDirent(baseTempDir)

	baseFilename, targetPath := getBackupPath(dest, course.GetID(), backupID)

	tempDir := filepath.Join(baseTempDir, baseFilename)
	err = db.DumpCourse(course, tempDir)
	if err != nil {
		return fmt.Errorf("Failed to dump course: '%w'.", err)
	}

	err = util.Zip(tempDir, targetPath, true)
	if err != nil {
		return fmt.Errorf("Failed to zip dumpped course dir '%s' into '%s': '%w'.", tempDir, targetPath, err)
	}

	return nil
}

func getBackupPath(dest string, basename string, backupID string) (string, string) {
	if backupID == "" {
		backupID = fmt.Sprintf("%d", timestamp.Now().ToMSecs())
	}

	offsetCount := 0
	baseFilename := fmt.Sprintf("%s-%s", basename, backupID)
	targetPath := filepath.Join(dest, baseFilename+".zip")

	for (targetPath == "") || (util.PathExists(targetPath)) {
		offsetCount++
		baseFilename = fmt.Sprintf("%s-%s-%d.zip", basename, backupID, offsetCount)
		targetPath = filepath.Join(dest, baseFilename+".zip")
	}

	return baseFilename, targetPath
}
