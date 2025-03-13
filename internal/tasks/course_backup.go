package tasks

import (
	"fmt"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/backup"
)

func RunCourseBackupTask(task *model.FullScheduledTask) error {
	if task.CourseID == "" {
		return fmt.Errorf("Course backup task has no course.")
	}

	err := backup.BackupCourse(task.CourseID)
	if err != nil {
		return fmt.Errorf("Failed to run course backup task: '%w'.", err)
	}

	return nil
}
