package model

import (
	"fmt"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

// The allowed types a task may have.
type TaskType string

const (
	TaskTypeUnknown TaskType = ""

	TaskTypeCourseBackup        TaskType = "backup"
	TaskTypeCourseEmailLogs     TaskType = "email-logs"
	TaskTypeCourseReport        TaskType = "report"
	TaskTypeCourseScoringUpload TaskType = "scoring-upload"
	TaskTypeCourseUpdate        TaskType = "update"
)

var taskTypeToString = map[TaskType]string{
	TaskTypeUnknown:             string(TaskTypeUnknown),
	TaskTypeCourseBackup:        string(TaskTypeCourseBackup),
	TaskTypeCourseEmailLogs:     string(TaskTypeCourseEmailLogs),
	TaskTypeCourseReport:        string(TaskTypeCourseReport),
	TaskTypeCourseScoringUpload: string(TaskTypeCourseScoringUpload),
	TaskTypeCourseUpdate:        string(TaskTypeCourseUpdate),
}

var stringToTaskType = map[string]TaskType{
	string(TaskTypeUnknown):             TaskTypeUnknown,
	string(TaskTypeCourseBackup):        TaskTypeCourseBackup,
	string(TaskTypeCourseEmailLogs):     TaskTypeCourseEmailLogs,
	string(TaskTypeCourseReport):        TaskTypeCourseReport,
	string(TaskTypeCourseScoringUpload): TaskTypeCourseScoringUpload,
	string(TaskTypeCourseUpdate):        TaskTypeCourseUpdate,
}

func (this TaskType) MarshalJSON() ([]byte, error) {
	return util.MarshalEnum(this, taskTypeToString)
}

func (this *TaskType) UnmarshalJSON(data []byte) error {
	value, err := util.UnmarshalEnum(data, stringToTaskType, true)
	if err == nil {
		*this = *value
	}

	return err
}

func validateTaskTypes(task *UserTaskInfo) error {
	switch task.Type {
	case TaskTypeCourseBackup:
		return nil
	case TaskTypeCourseReport:
		return validateTaskTypeCourseReport(task)
	case TaskTypeCourseScoringUpload:
		return nil
	case TaskTypeCourseUpdate:
		return nil
	case TaskTypeCourseEmailLogs:
		return validateTaskTypeCourseEmailLogs(task)
	default:
		return fmt.Errorf("Unknown task type: '%s'.", task.Type)
	}
}

func validateTaskTypeCourseEmailLogs(task *UserTaskInfo) error {
	err := validateEmailList(task)
	if err != nil {
		return err
	}

	rawQuery, err := GetTaskOptionAsType(task, "query", log.RawLogQuery{})
	if err != nil {
		return fmt.Errorf("Unable to parse 'query' key: '%w'.", err)
	}

	_, err = rawQuery.ParseJoin()
	if err != nil {
		return err
	}

	task.Options["query"] = rawQuery

	task.Options["send-empty"] = (task.Options["send-empty"] == true)

	return nil
}

func validateTaskTypeCourseReport(task *UserTaskInfo) error {
	return validateEmailList(task)
}

func validateEmailList(task *UserTaskInfo) error {
	to, err := GetTaskOptionAsType(task, "to", make([]string, 0))
	if err != nil {
		return fmt.Errorf("'to' value is not properly formatted: '%w'.", err)
	}

	task.Options["to"] = to

	if !task.Disabled && (len(to) == 0) {
		return fmt.Errorf("Task is not disabled, but no email recipients are declared in the 'to' value.")
	}

	return nil
}
