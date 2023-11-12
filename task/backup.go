package task

import (
    "fmt"
    "path/filepath"
    "time"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
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

    return RunBackup(course, task.Dest);
}

// Perform a backup.
// If dest is not specified, it will be picked up from config.BACKUP_DIR.
func RunBackup(course model.Course, dest string) error {
    if (dest == "") {
        dest = config.BACKUP_DIR.Get();
    }

    err := util.MkDir(dest);
    if (err != nil) {
        return fmt.Errorf("Could not create dest dir '%s': '%w'.", dest, err);
    }

    baseTempDir, err := util.MkDirTemp("autograder-backup-course-");
    if (err != nil) {
        return fmt.Errorf("Could not create temp backup dir: '%w'.", err);
    }
    defer util.RemoveDirent(baseTempDir);

    baseFilename, targetPath := getBackupPath(dest, course.GetID());

    tempDir := filepath.Join(baseTempDir, baseFilename);
    err = db.DumpCourse(course, tempDir);
    if (err != nil) {
        return fmt.Errorf("Failed to dump course: '%w'.", err);
    }

    err = util.Zip(tempDir, targetPath, true);
    if (err != nil) {
        return fmt.Errorf("Failed to zip dumpped course dir '%s' into '%s': '%w'.", tempDir, targetPath, err);
    }

    return nil;
}

func getBackupPath(dest string, basename string) (string, string) {
    backupID := time.Now().Unix();
    offsetCount := 0;
    baseFilename := fmt.Sprintf("%s-%d", basename, backupID);
    targetPath := filepath.Join(dest, baseFilename + ".zip");

    for ((targetPath == "") || (util.PathExists(targetPath))) {
        offsetCount++;
        baseFilename = fmt.Sprintf("%s-%d-%d.zip", basename, backupID, offsetCount);
        targetPath = filepath.Join(dest, baseFilename + ".zip");
    }

    return baseFilename, targetPath;
}
