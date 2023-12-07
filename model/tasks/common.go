// Model for scheduled tasks (general and specific).
// Does not include any code for running scheduled tasks.

package tasks

type TaskCourse interface {
    GetID() string
    GetAssignmentLMSIDs() ([]string, []string)
    HasLMSAdapter() bool
}

type ScheduledTask interface {
    IsDisabled() bool
    GetTime() *ScheduledTime
    String() string
    Validate(TaskCourse) error
}
