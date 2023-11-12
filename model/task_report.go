package model

import (
    "fmt"

    "github.com/eriq-augustine/autograder/config"
)

type ReportTask struct {
    Disable bool `json:"disable"`
    When ScheduledTime `json:"when"`
    To []string `json:"to"`

    CourseID string `json:"-"`
}

func (this *ReportTask) Validate(course Course) error {
    this.When.id = fmt.Sprintf("report-%s", course.GetID());
    this.CourseID = course.GetID();

    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.Disable = (this.Disable || config.NO_TASKS.Get());

    if (!this.Disable && (len(this.To) == 0)) {
        return fmt.Errorf("Report task is not disabled, but no email recipients are declared.");
    }

    return nil;
}

func (this *ReportTask) IsDisabled() bool {
    return this.Disable;
}

func (this *ReportTask) GetTime() *ScheduledTime {
    return &this.When;
}

func (this *ReportTask) String() string {
    return fmt.Sprintf("Report on course '%s': '%s'.",
            this.CourseID, this.When.String());
}
