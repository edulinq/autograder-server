package model

import (
	"github.com/edulinq/autograder/internal/util"
)

// Where a task originated from.
type TaskSource string

const (
	TaskSourceUnknown TaskSource = ""
	TaskSourceCourse             = "course"
)

var taskSourceToString = map[TaskSource]string{
	TaskSourceUnknown: string(TaskSourceUnknown),
	TaskSourceCourse:  string(TaskSourceCourse),
}

var stringToTaskSource = map[string]TaskSource{
	string(TaskSourceUnknown): TaskSourceUnknown,
	string(TaskSourceCourse):  TaskSourceCourse,
}

func (this TaskSource) MarshalJSON() ([]byte, error) {
	return util.MarshalEnum(this, taskSourceToString)
}

func (this *TaskSource) UnmarshalJSON(data []byte) error {
	value, err := util.UnmarshalEnum(data, stringToTaskSource, true)
	if err == nil {
		*this = *value
	}

	return err
}
