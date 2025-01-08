package tasks

import (
	"fmt"
)

type ScoringUploadTask struct {
	*BaseTask

	DryRun bool `json:"dry-run"`
}

func (this *ScoringUploadTask) Validate(course TaskCourse) error {
	this.BaseTask.Name = "scoring"

	err := this.BaseTask.Validate(course)
	if err != nil {
		return err
	}

	if !course.HasLMSAdapter() {
		return fmt.Errorf("Score and Upload task course must have an LMS adapter.")
	}

	return nil
}
