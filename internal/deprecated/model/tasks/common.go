package tasks

import (
	"github.com/edulinq/autograder/internal/util"
)

type ScheduledTask interface {
	IsDisabled() bool
	GetTimes() []*util.ScheduledTime
}

type BaseTask struct {
	Disable bool                  `json:"disable"`
	When    []*util.ScheduledTime `json:"when"`
}

func (this *BaseTask) IsDisabled() bool {
	return this.Disable
}

func (this *BaseTask) GetTimes() []*util.ScheduledTime {
	return this.When
}
