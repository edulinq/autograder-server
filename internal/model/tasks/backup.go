package tasks

import (
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/util"
)

type BackupTask struct {
	*BaseTask

	Dest     string `json:"-"`
	BackupID string `json:"-"`
}

func (this *BackupTask) Validate(course TaskCourse) error {
	this.BaseTask.Name = "backup"

	err := this.BaseTask.Validate(course)
	if err != nil {
		return err
	}

	// For backup location check (in order): Dest, config.TASK_BACKUP_DIR, config.GetBackupDir().
	// The latter two are checked in config.GetTaskBackupDir().
	if this.Dest == "" {
		this.Dest = config.GetTaskBackupDir()
	}

	if util.IsFile(this.Dest) {
		return fmt.Errorf("Backup directory exists and is a file: '%s'.", this.Dest)
	}

	return nil
}
