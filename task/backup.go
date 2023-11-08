package task

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

func RunBackupTask(course model.Course, rawTask model.ScheduledTask) error {
    task, ok := rawTask.(*model.BackupTask);
    if (!ok) {
        return fmt.Errorf("Task is not a BackupTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return nil;
    }

    return RunBackup(task.Source, task.Dest, task.Basename);
}

// Perform a backup.
// If dest is not specified, it will be picked up from config.BACKUP_DIR.
func RunBackup(source string, dest string, basename string) error {
    if (dest == "") {
        dest = config.BACKUP_DIR.Get();
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
    err := util.Zip(source, targetPath, false);
    if (err != nil) {
        log.Debug().Str("source", source).Str("dest", dest).Str("basename", basename).Msg("Backup failed.");
        return err;
    }

    log.Debug().Str("source", source).Str("dest", dest).Str("basename", basename).Msg("Backup completed sucessfully.");
    return nil;
}
