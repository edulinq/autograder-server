// Model for scheduled tasks (general and specific).
// Does not include any code for running scheduled tasks.

package tasks

import (
    "fmt"
    "strings"

    "github.com/eriq-augustine/autograder/config"
)

type TaskCourse interface {
    GetID() string
    GetAssignmentLMSIDs() ([]string, []string)
    HasLMSAdapter() bool
}

type ScheduledTask interface {
    GetID() string
    GetCourseID() string
    IsDisabled() bool
    GetTimes() []*ScheduledTime
    String() string
    Validate(TaskCourse) error
}

type BaseTask struct {
    Disable bool `json:"disable"`
    When []*ScheduledTime `json:"when"`

    ID string `json:"-"`
    Name string `json:"-"`
    CourseID string `json:"-"`
}

func (this *BaseTask) GetID() string {
    return this.ID;
}

func (this *BaseTask) GetCourseID() string {
    return this.CourseID;
}

func (this *BaseTask) IsDisabled() bool {
    return this.Disable;
}

func (this *BaseTask) GetTimes() []*ScheduledTime {
    return this.When;
}

func (this *BaseTask) String() string {
    times := make([]string, 0, len(this.When));
    for _, when := range this.When {
        times = append(times, when.String());
    }

    disabled := "";
    if (this.Disable) {
        disabled = " (disabled) ";
    }

    return fmt.Sprintf("Task (%s) %s scheduled for [%s]",
        this.Name,
        disabled,
        strings.Join(times, ", "));
}

func (this *BaseTask) Validate(course TaskCourse) error {
    if (this.Name == "") {
        return fmt.Errorf("No name provided to the task.");
    }

    this.CourseID = course.GetID();
    this.ID = course.GetID() + "::" + this.Name;

    for i, when := range this.When {
        if (when == nil) {
            return fmt.Errorf("%d when instance is nil.", i);
        }

        err := when.Validate();
        if (err != nil) {
            return fmt.Errorf("Failed to validate %d when instance: '%w'.", i, err);
        }
    }

    this.Disable = (this.Disable || config.NO_TASKS.Get());

    return nil;
}
