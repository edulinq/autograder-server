package model

import (
	dtasks "github.com/edulinq/autograder/internal/deprecated/model/tasks"
)

func DeprecatedTaskToStandard(deprecatedTask dtasks.ScheduledTask) []*UserTaskInfo {
	tasks := make([]*UserTaskInfo, 0)

	for _, when := range deprecatedTask.GetTimes() {
		task := &UserTaskInfo{
			Disable: deprecatedTask.IsDisabled(),
			When:    when,
		}

		taskType := TaskTypeUnknown
		switch oldTask := deprecatedTask.(type) {
		case *dtasks.BackupTask:
			taskType = TaskTypeCourseBackup
		case *dtasks.ReportTask:
			taskType = TaskTypeCourseReport
			task.Options["to"] = oldTask.To
		case *dtasks.ScoringUploadTask:
			taskType = TaskTypeCourseScoringUpload
		case *dtasks.CourseUpdateTask:
			taskType = TaskTypeCourseUpdate
		case *dtasks.EmailLogsTask:
			taskType = TaskTypeEmailLogs
			task.Options["to"] = oldTask.To
			task.Options["query"] = oldTask.RawQuery
			task.Options["send-empty"] = oldTask.SendEmpty
		}

		task.Type = taskType

		tasks = append(tasks, task)
	}

	return tasks
}
