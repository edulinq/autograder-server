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
    Basename string `json:"-"`
    Source string `json:"-"`
    Dest string `json:"-"`
}

func (this *BackupTask) Validate(course Course) error {
    this.When.id = fmt.Sprintf("backup-%s", course.GetID());
    this.CourseID = course.GetID();

    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.Disable = (this.Disable || config.NO_TASKS.Get());

    this.Basename = course.GetID();
    if (this.Basename == "") {
        return fmt.Errorf("Backup basename cannot be empty.");
    }

    this.Source = course.GetSourceDir();
    if (!util.PathExists(this.Source)) {
        return fmt.Errorf("Backup source path '%s' does not exist.", this.Source);
    }

    this.Dest = config.BACKUP_DIR.Get();
    if (util.IsFile(this.Dest)) {
        return fmt.Errorf("Backup directory exists and is a file: '%s'.", this.Dest);
    }

    return nil;
}

func (this *BackupTask) IsDisabled() bool {
    return this.Disable;
}

func (this *BackupTask) GetTime() ScheduledTime {
    return this.When;
}

func (this *BackupTask) String() string {
    return fmt.Sprintf("Backup of course '%s': '%s' to '%s' at '%s' (next time: '%s').",
            this.CourseID, this.Source, this.Dest, this.When.String(), this.When.ComputeNext());
}
