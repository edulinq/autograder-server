package tasks

import (
    "fmt"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

type BackupTask struct {
    *BaseTask

    Dest string `json:"-"`
    BackupID string `json:"-"`
}

func (this *BackupTask) Validate(course TaskCourse) error {
    this.BaseTask.Name = "backup";

    err := this.BaseTask.Validate(course);
    if (err != nil) {
        return err;
    }

    this.Dest = config.GetBackupDir();
    if (util.IsFile(this.Dest)) {
        return fmt.Errorf("Backup directory exists and is a file: '%s'.", this.Dest);
    }

    return nil;
}
