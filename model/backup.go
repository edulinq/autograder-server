package model

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

type BackupTask struct {
    Disable bool `json:"disable"`
    When ScheduledTime `json:"when"`

    basename string `json:"-"`
    source string `json:"-"`
    dest string `json:"-"`
}

func (this *BackupTask) Validate(source string, basename string) error {
    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.basename = basename;
    if (this.basename == "") {
        return fmt.Errorf("Backup basename cannot be empty.");
    }

    this.source = source;
    if (!util.PathExists(this.source)) {
        return fmt.Errorf("Backup source path '%s' does not exist.", this.source);
    }

    this.dest = config.BACKUP_DIR.GetString();
    if (util.IsFile(this.dest)) {
        return fmt.Errorf("Backup directory exists and is a file: '%s'.", this.dest);
    }

    return nil;
}

func (this *BackupTask) String() string {
    return fmt.Sprintf("Backup '%s' to '%s' at '%s' (next time: '%s').", this.source, this.dest, this.When.String(), this.When.ComputeNext());
}

// Schedule this task to be regularly run at the scheduled time.
func (this *BackupTask) Schedule() {
    this.When.Schedule(func() {
        err := this.Run();
        if (err != nil) {
            log.Error().Err(err).Str("source", this.source).Str("dest", this.dest).Msg("Backup task failed.");
        }
    });
}

// Stop any scheduled executions of this task.
func (this *BackupTask) Stop() {
    this.When.Stop();
}

// Run the task regardless of schedule.
func (this *BackupTask) Run() error {
    return RunBackup(this.source, this.dest, this.basename);
}

// Do a backup without an attatched object.
// If dest is not specified, it will be picked up from config.BACKUP_DIR.
func RunBackup(source string, dest string, basename string) error {
    if (dest == "") {
        dest = config.BACKUP_DIR.GetString();
    }

    os.MkdirAll(dest, 0755);

    backupID := time.Now().Unix();
    offsetCount := 0;
    targetPath := filepath.Join(dest, fmt.Sprintf("%s-%d.zip", basename, backupID));

    for ((targetPath == "") || (util.PathExists(targetPath))) {
        offsetCount++;
        targetPath = filepath.Join(dest, fmt.Sprintf("%s-%d-%d.zip", basename, backupID, offsetCount));
    }

    log.Debug().Str("source", source).Str("dest", dest).Str("basename", basename).Msg("Starting backup.");
    err := util.Zip(source, targetPath);
    if (err != nil) {
        log.Debug().Str("source", source).Str("dest", dest).Str("basename", basename).Msg("Backup failed.");
        return err;
    }

    log.Debug().Str("source", source).Str("dest", dest).Str("basename", basename).Msg("Backup completed sucessfully.");
    return nil;
}
