package model

import (
    "fmt"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

type BackupTask struct {
    Disable bool `json:"disable"`
    When ScheduledTime `json:"when"`

    CourseID string `json:"-"`
    Dest string `json:"-"`
}

func (this *BackupTask) Validate(course *Course) error {
    this.When.id = fmt.Sprintf("backup-%s", course.GetID());
    this.CourseID = course.GetID();

    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.Disable = (this.Disable || config.NO_TASKS.Get());

    this.Dest = config.BACKUP_DIR.Get();
    if (util.IsFile(this.Dest)) {
        return fmt.Errorf("Backup directory exists and is a file: '%s'.", this.Dest);
    }

    return nil;
}

func (this *BackupTask) IsDisabled() bool {
    return this.Disable;
}

func (this *BackupTask) GetTime() *ScheduledTime {
    return &this.When;
}

func (this *BackupTask) String() string {
    return fmt.Sprintf("Backup of course '%s' to '%s' at '%s'.",
            this.CourseID, this.Dest, this.When.String());
}
