package model

type ScheduledTask interface {
    Schedule()
}

type ScheduledCourseTask interface {
    ScheduledTask
    Validate(*Course) error
}

