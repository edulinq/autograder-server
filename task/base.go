package task

import (
    "github.com/eriq-augustine/autograder/canvas"
)

type TaskCourseSource interface {
    GetID() string
    GetSourceDir() string
    GetCanvasInstanceInfo() *canvas.CanvasInstanceInfo
    // (canvas ids, assignment ids)
    GetCanvasIDs() ([]string, []string)
    FullScoringAndUpload(bool) error
}

type ScheduledTask interface {
    Schedule()
}

type ScheduledCourseTask interface {
    ScheduledTask
    Validate(TaskCourseSource) error
}
