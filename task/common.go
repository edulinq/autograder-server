// Implementations for running tasks.
package task

import (
    "fmt"
    "time"

    "github.com/rs/zerolog/log"
    "golang.org/x/exp/maps"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/model/tasks"
)

// {courseID: [taskTimer, ...], ...}
var courseTimers map[string][]*time.Timer = make(map[string][]*time.Timer);

type RunFunc func(*model.Course, tasks.ScheduledTask) error;

func Schedule(course *model.Course, target tasks.ScheduledTask) error {
    if (target.IsDisabled() || config.NO_TASKS.Get()) {
        return nil;
    }

    var runFunc RunFunc;
    switch target.(type) {
        case *tasks.BackupTask:
            runFunc = RunBackupTask;
        case *tasks.ReportTask:
            runFunc = RunReportTask;
        case *tasks.ScoringUploadTask:
            runFunc = RunScoringUploadTask;
        default:
            return fmt.Errorf("Unknown task type: %t (%v).", target, target);
    }

    nextRunTime := target.GetTime().ComputeNext();
    timer := time.AfterFunc(nextRunTime.Sub(time.Now()), func() {
        runTask(course.GetID(), target, runFunc);
    });

    courseTimers[course.GetID()] = append(courseTimers[course.GetID()], timer);

    log.Debug().Str("task", target.String()).Any("next-time", nextRunTime).Msg("Task scheduled.");
    return nil;
}

func runTask(courseID string, target tasks.ScheduledTask, runFunc RunFunc) {
    log.Debug().Str("task", target.String()).Msg("Task started.");

    course, err := db.GetCourse(courseID);
    if (err != nil) {
        log.Error().Err(err).Str("course-id", courseID).Str("task", target.String()).Msg("Failed to get course for task.");
        return;
    }

    if (course == nil) {
        log.Error().Str("course-id", courseID).Str("task", target.String()).Msg("Could not find course for task.");
        return;
    }

    err = target.Validate(course);
    if (err != nil) {
        log.Error().Err(err).Str("course-id", courseID).Str("task", target.String()).Msg("Task failed validation.");
        return;
    }

    err = runFunc(course, target);
    if (err != nil) {
        log.Error().Err(err).Str("course-id", courseID).Str("task", target.String()).Msg("Task run failed.");
        return;
    }

    log.Debug().Str("course-id", courseID).Str("task", target.String()).Msg("Task finished.");
}

func StopCourse(courseID string) {
    for _, timer := range courseTimers[courseID] {
        if (!timer.Stop()) {
            // Clear the channel.
            <- timer.C;
        }
    }

    delete(courseTimers, courseID);
}

func StopAll() {
    keys := maps.Keys(courseTimers);
    for _, key := range keys {
        StopCourse(key);
    }
}
