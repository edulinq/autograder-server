package tasks

import (
	"github.com/edulinq/autograder/internal/log"
)

type BackupTask struct {
	*BaseTask

	Dest     string `json:"-"`
	BackupID string `json:"-"`
}

type CourseUpdateTask struct {
	*BaseTask
}

type EmailLogsTask struct {
	*BaseTask

	To        []string `json:"to"`
	SendEmpty bool     `json:"send-empty"`

	RawQuery log.RawLogQuery `json:"query"`
}

type ReportTask struct {
	*BaseTask

	To []string `json:"to"`
}

type ScoringUploadTask struct {
	*BaseTask

	DryRun bool `json:"dry-run"`
}
