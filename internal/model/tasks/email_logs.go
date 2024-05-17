package tasks

import (
	"fmt"

	"github.com/edulinq/autograder/internal/common"
)

type EmailLogsTask struct {
	*BaseTask

	To        []string `json:"to"`
	SendEmpty bool     `json:"send-empty"`

	common.RawLogQuery
}

func (this *EmailLogsTask) Validate(course TaskCourse) error {
	this.BaseTask.Name = "email-logs"

	err := this.BaseTask.Validate(course)
	if err != nil {
		return err
	}

	if !this.Disable && (len(this.To) == 0) {
		return fmt.Errorf("EmailLogs task is not disabled, but no email recipients are declared.")
	}

	_, err = this.RawLogQuery.ParseJoin(course)
	if err != nil {
		return err
	}

	return nil
}
