package task

import (
    "github.com/eriq-augustine/autograder/lms/adapter"
    "github.com/eriq-augustine/autograder/report"
)

type TaskCourseSource interface {
    report.ReportingSources

    GetID() string
    GetSourceDir() string
    GetLMSAdapter() *adapter.LMSAdapter
    // (LMS ids, assignment ids)
    GetAssignmentLMSIDs() ([]string, []string)
    FullScoringAndUpload(bool) error
}

type ScheduledTask interface {
    Schedule()
}

type ScheduledCourseTask interface {
    ScheduledTask
    Validate(TaskCourseSource) error
}
