package tasks

import (
	"fmt"
)

type ReportTask struct {
	*BaseTask

	To []string `json:"to"`
}

func (this *ReportTask) Validate(course TaskCourse) error {
	this.BaseTask.Name = "report"

	err := this.BaseTask.Validate(course)
	if err != nil {
		return err
	}

	if !this.Disable && (len(this.To) == 0) {
		return fmt.Errorf("Report task is not disabled, but no email recipients are declared.")
	}

	return nil
}
