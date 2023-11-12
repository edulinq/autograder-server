package model

type ScheduledTask interface {
    IsDisabled() bool
    GetTime() *ScheduledTime
    String() string
    Validate(Course) error
}
