package model

import (
	"fmt"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type FullScheduledTask struct {
	UserTaskInfo
	SystemTaskInfo
}

// Information about a task supplied by the user.
type UserTaskInfo struct {
	Type    TaskType              `json:"type"`
	Name    string                `json:"name,omitempty"`
	Disable bool                  `json:"disable,omitempty"`
	When    *common.ScheduledTime `json:"when,omitempty"`
	Options map[string]any        `json:"options,omitempty"`
}

// Information about a task supplied by the autograder.
type SystemTaskInfo struct {
	Source       TaskSource          `json:"source"`
	LastRunTime  timestamp.Timestamp `json:"next-runtime"`
	NextRunTime  timestamp.Timestamp `json:"next-runtime"`
	Hash         string              `json:"hash"`
	CourseID     string              `json:"course-id,omitempty"`
	AssignmentID string              `json:"assignment-id,omitempty"`
	UserEmail    string              `json:"user-email,omitempty"`
}

func (this *UserTaskInfo) String() string {
	name := ""
	if this.Name != "" {
		name = fmt.Sprintf(" (%s)", this.Name)
	}

	timeString := "never"
	if this.When != nil {
		timeString = this.When.String()
	}

	disabled := " "
	if this.Disable {
		disabled = " (disabled) "
	}

	return fmt.Sprintf("Task%s%sof type '%s' scheduled for [%s]", name, disabled, this.Type, timeString)
}

func (this *UserTaskInfo) Validate() error {
	if this == nil {
		return fmt.Errorf("Nil tasks are not allowed.")
	}

	if (this.When == nil) && (!this.Disable) {
		return fmt.Errorf("Scheduled time to run ('when') is not supplied and the task is not disabled.")
	}

	if this.When != nil {
		err := this.When.Validate()
		if err != nil {
			return fmt.Errorf("Failed to validate scheduled time to run: '%w'.", err)
		}
	}

	if this.Options == nil {
		this.Options = make(map[string]any, 0)
	}

	return validateTaskTypes(this)
}

func (this *UserTaskInfo) ToFullCourseTask(courseID string) (*FullScheduledTask, error) {
	hash, err := util.Sha256HashFromJSONObject(this)
	if err != nil {
		return nil, fmt.Errorf("Unable to make hash from task: '%w'.", err)
	}

	systemTaskInfo := SystemTaskInfo{
		Source:      TaskSourceCourse,
		LastRunTime: timestamp.Zero(),
		NextRunTime: this.When.ComputeNextTimeFromNow(),
		Hash:        hash,
		CourseID:    courseID,
	}

	err = systemTaskInfo.Validate()
	if err != nil {
		return nil, fmt.Errorf("Failed to validate system task info: '%w'.", err)
	}

	fullTask := &FullScheduledTask{
		UserTaskInfo:   *this,
		SystemTaskInfo: systemTaskInfo,
	}

	return fullTask, fullTask.Validate()
}

func (this *SystemTaskInfo) Validate() error {
	if this.Hash == "" {
		return fmt.Errorf("Hash cannot be empty.")
	}

	var err error

	if this.CourseID != "" {
		this.CourseID, err = common.ValidateID(this.CourseID)
		if err != nil {
			return fmt.Errorf("Course ID is not valid: '%w'.", err)
		}
	}

	if this.AssignmentID != "" {
		this.AssignmentID, err = common.ValidateID(this.AssignmentID)
		if err != nil {
			return fmt.Errorf("Assignment ID is not valid: '%w'.", err)
		}
	}

	return nil
}

func (this *FullScheduledTask) Validate() error {
	err := this.UserTaskInfo.Validate()
	if err != nil {
		return err
	}

	return this.SystemTaskInfo.Validate()
}

// Merge times according to task updating logic
// (as if a new task (this) was just read in and it replacing the exiting task (oldTask)).
func (this *FullScheduledTask) MergeTimes(oldTask *FullScheduledTask) {
	if oldTask == nil {
		return
	}

	// Always take the last run time from the old task.
	this.LastRunTime = oldTask.LastRunTime

	// Take the sooner of the next run times.
	if this.NextRunTime > oldTask.NextRunTime {
		this.NextRunTime = oldTask.NextRunTime
	}
}

func GetTaskOptionAsType[T any](task *UserTaskInfo, key string, defaultValue T) (T, error) {
	rawValue, ok := task.Options[key]
	if !ok {
		return defaultValue, nil
	}

	return util.JSONTransformTypes(rawValue, defaultValue)
}
