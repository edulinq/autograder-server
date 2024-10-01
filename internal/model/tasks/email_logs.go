package tasks

import (
	"fmt"

	"github.com/edulinq/autograder/internal/log"
)

type EmailLogsTask struct {
	*BaseTask

	To        []string `json:"to"`
	SendEmpty bool     `json:"send-empty"`

	log.RawLogQuery
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

	_, err = this.RawLogQuery.ParseJoin()
	if err != nil {
		return err
	}

	return nil
}

func (this *EmailLogsTask) String() string {
	return this.BaseTask.String()
}
