// Model for scheduled tasks (general and specific).
// Does not include any code for running scheduled tasks.

package tasks

import (
	"fmt"
	"strings"
	"sync"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
)

type TaskCourse interface {
	GetID() string
	GetAssignmentLMSIDs() ([]string, []string)
	HasAssignment(id string) bool
	HasLMSAdapter() bool
}

type ScheduledTask interface {
	GetID() string
	GetCourseID() string
	IsDisabled() bool
	GetTimes() []*common.ScheduledTime
	// Get the minimum time between task runs.
	// The boolean return will be false if there are no times (infinite durtion).
	GetMinDurationMS() (int64, bool)
	String() string
	Validate(TaskCourse) error
	GetLock() *sync.Mutex
}

type BaseTask struct {
	Disable bool                    `json:"disable"`
	When    []*common.ScheduledTime `json:"when"`

	ID       string `json:"-"`
	Name     string `json:"-"`
	CourseID string `json:"-"`
	// A lock to ensure only one instance of the task is runnning at a time.
	Lock *sync.Mutex `json:"-"`
}

func (this *BaseTask) GetID() string {
	return this.ID
}

func (this *BaseTask) GetCourseID() string {
	return this.CourseID
}

func (this *BaseTask) IsDisabled() bool {
	return this.Disable
}

func (this *BaseTask) GetTimes() []*common.ScheduledTime {
	return this.When
}

func (this *BaseTask) GetMinDurationMS() (int64, bool) {
	if len(this.When) == 0 {
		return 0, false
	}

	var minDuration int64 = -1
	for _, when := range this.When {
		duration := when.TotalMSecs()
		if (minDuration < 0) || (duration < minDuration) {
			minDuration = duration
		}
	}

	return minDuration, true
}

func (this *BaseTask) String() string {
	times := make([]string, 0, len(this.When))
	for _, when := range this.When {
		times = append(times, when.String())
	}

	disabled := ""
	if this.Disable {
		disabled = " (disabled) "
	}

	return fmt.Sprintf("Task (%s) %s scheduled for [%s]",
		this.Name,
		disabled,
		strings.Join(times, ", "))
}

func (this *BaseTask) Validate(course TaskCourse) error {
	this.Lock = &sync.Mutex{}

	if this.Name == "" {
		return fmt.Errorf("No name provided to the task.")
	}

	this.CourseID = course.GetID()
	this.ID = course.GetID() + "::" + this.Name

	for i, when := range this.When {
		if when == nil {
			return fmt.Errorf("%d when instance is nil.", i)
		}

		err := when.Validate()
		if err != nil {
			return fmt.Errorf("Failed to validate %d when instance: '%w'.", i, err)
		}
	}

	this.Disable = (this.Disable || config.NO_TASKS.Get())

	return nil
}

func (this *BaseTask) GetLock() *sync.Mutex {
	return this.Lock
}
