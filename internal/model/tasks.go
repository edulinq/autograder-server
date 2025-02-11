package model

import (
	"fmt"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type FullScheduledTask struct {
	UserTaskInfo
	SystemTaskInfo
}

// Information about a task supplied by the user.
type UserTaskInfo struct {
	Type     TaskType            `json:"type"`
	Name     string              `json:"name,omitempty"`
	Disabled bool                `json:"disabled,omitempty"`
	When     *util.ScheduledTime `json:"when,omitempty"`
	Options  map[string]any      `json:"options,omitempty"`
}

// Information about a task supplied by the autograder.
type SystemTaskInfo struct {
	Source       TaskSource          `json:"source"`
	LastRunTime  timestamp.Timestamp `json:"last-runtime"`
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
	if this.Disabled {
		disabled = " (disabled) "
	}

	return fmt.Sprintf("Task%s%sof type '%s' scheduled for [%s]", name, disabled, this.Type, timeString)
}

func (this *UserTaskInfo) Validate() error {
	if this == nil {
		return fmt.Errorf("Nil tasks are not allowed.")
	}

	if (this.When == nil) && (!this.Disabled) {
		return fmt.Errorf("Scheduled time to run ('when') is not supplied and the task is not disabled.")
	}

	if this.When != nil {
		err := this.When.Validate()
		if err != nil {
			return fmt.Errorf("Failed to validate scheduled time to run: '%w'.", err)
		}

		minPeriodMSecs := int64(config.TASK_MIN_PERIOD_SECS.Get() * 1000)
		if this.When.TotalMSecs() < minPeriodMSecs {
			return fmt.Errorf("Task is scheduled too often. Min Period (msecs): %d, Current Period (msecs): %d.", minPeriodMSecs, this.When.TotalMSecs())
		}
	}

	if this.Options == nil {
		this.Options = make(map[string]any, 0)
	}

	return validateTaskTypes(this)
}

// Create a full task from user-defined information.
// The created hash will be consistent as long as the user-defined information stays the same.
// Will return nil if the task is disabled.
func (this *UserTaskInfo) ToFullCourseTask(courseID string) (*FullScheduledTask, error) {
	if this.Disabled {
		return nil, nil
	}

	configHash, err := util.Sha256HashFromJSONObject(this)
	if err != nil {
		return nil, fmt.Errorf("Unable to make hash from task: '%w'.", err)
	}

	// Computing the next time depends on how the task is to be repeated: daily or periodically.
	// Daily tasks will be computed starting at right now, so they will run the next chosen time period.
	// Periodic ("every") tasks will be computed starting at zero time, so they will get a run in right away.
	// However, if a periodic task has an existing run in the DB, then it will be merged and this time overwritten (see MergeTimes()).
	// Daily tasks will also get merged, but the two tasks should share the same next run time.
	baselineTime := timestamp.Zero()
	if !this.When.Daily.IsEmpty() {
		baselineTime = timestamp.Now()
	}

	systemTaskInfo := SystemTaskInfo{
		Source:      TaskSourceCourse,
		LastRunTime: timestamp.Zero(),
		NextRunTime: this.When.ComputeNextTime(baselineTime),
		// Use a combination of the course id, type, and hashed user config as the hash.
		Hash:     util.JoinStrings("::", courseID, string(this.Type), configHash),
		CourseID: courseID,
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
	if (this == nil) || (oldTask == nil) {
		return
	}

	// Always take the last run time from the old task.
	this.LastRunTime = oldTask.LastRunTime

	// For daily tasks, take the earlier time.
	// For periodic ("every") tasks, take the latter time.
	// Note that daily tasks should share the same next run time,
	// but a loading a task (e.g., from a course) at the exact right time could cause a skip if we took the latter time.

	if !this.When.Daily.IsEmpty() {
		if this.NextRunTime > oldTask.NextRunTime {
			this.NextRunTime = oldTask.NextRunTime
		}
	} else {
		if this.NextRunTime < oldTask.NextRunTime {
			this.NextRunTime = oldTask.NextRunTime
		}
	}
}

// Advance the run times as if this task successfully completed a run.
func (this *FullScheduledTask) AdvanceRunTimes() {
	this.LastRunTime = timestamp.Now()
	this.NextRunTime = this.When.ComputeNextTimeFromNow()
}

func (this *FullScheduledTask) LogValue() []*log.Attr {
	attrs := []*log.Attr{
		log.NewAttr("task-type", this.Type),
	}

	if this.Name != "" {
		attrs = append(attrs, log.NewAttr("name", this.Name))
	}

	if this.CourseID != "" {
		attrs = append(attrs, log.NewCourseAttr(this.CourseID))
	}

	if this.AssignmentID != "" {
		attrs = append(attrs, log.NewAssignmentAttr(this.AssignmentID))
	}

	if this.UserEmail != "" {
		attrs = append(attrs, log.NewUserAttr(this.UserEmail))
	}

	return attrs
}

func GetTaskOptionAsType[T any](task *UserTaskInfo, key string, defaultValue T) (T, error) {
	rawValue, ok := task.Options[key]
	if !ok {
		return defaultValue, nil
	}

	return util.JSONTransformTypes(rawValue, defaultValue)
}
