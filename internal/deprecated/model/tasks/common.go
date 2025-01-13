package tasks

import (
	"github.com/edulinq/autograder/internal/common"
)

type ScheduledTask interface {
	IsDisabled() bool
	GetTimes() []*common.ScheduledTime
}

type BaseTask struct {
	Disable bool                    `json:"disable"`
	When    []*common.ScheduledTime `json:"when"`
}

func (this *BaseTask) IsDisabled() bool {
	return this.Disable
}

func (this *BaseTask) GetTimes() []*common.ScheduledTime {
	return this.When
}
