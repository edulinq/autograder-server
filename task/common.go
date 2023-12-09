// Implementations for running tasks.
package task

import (
    "fmt"
    "time"
    "sync"

    "github.com/rs/zerolog/log"
    "golang.org/x/exp/maps"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/model/tasks"
)

var timersLock sync.Mutex;

type timerInfo struct {
    ID string
    CourseID string
    Timer *time.Timer
    Lock *sync.Mutex
    // If this specific timer instance has been stopped, do not run the task.
    // This catches a different case than stoppedTasks
    // (which prevents a task from re-scheduling).
    Stopped bool
}

// {courseID: {ID: timerInfo, ...}, ...}
var timers map[string]map[string]*timerInfo = make(map[string]map[string]*timerInfo);

// Stopped tasks should not be rescheduled automatically.
// Manually scheduling a stopped task (via Schedule()) will remove it from this map.
// The key (id) is the same as timerInfo.ID.
var stoppedTasks map[string]bool = make(map[string]bool);

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
        case *tasks.TestTask:
            runFunc = RunTestTask;
        default:
            return fmt.Errorf("Unknown task type: %t (%v).", target, target);
    }

    for i, when := range target.GetTimes() {
        // ID unique to every (task, timer).
        id := fmt.Sprintf("%s::%03d", target.GetID(), i);

        timersLock.Lock();
        delete(stoppedTasks, id);
        timersLock.Unlock();

        err := scheduleTask(course.GetID(), target, id, runFunc, when);
        if (err != nil) {
            return fmt.Errorf("Failed to schedule task (%s): '%w'.", target.String(), err);
        }
    }

    return nil;
}

func scheduleTask(courseID string, target tasks.ScheduledTask, id string, runFunc RunFunc, when *tasks.ScheduledTime) error {
    timersLock.Lock();
    defer timersLock.Unlock();

    // Ensure this task has not been stopped,
    // and therefore should not be scheduled.
    stopped, exists := stoppedTasks[id];
    if (exists && stopped) {
        return nil;
    }

    // In addition to the general timers lock, create a lock just for this timer that prevents
    // a scheduled task from calling too soon (before the timerInfo is setup).
    taskLock := &sync.Mutex{};
    taskLock.Lock();
    defer taskLock.Unlock();

    nextRunTime := when.ComputeNextTimeFromNow();
    timer := time.AfterFunc(nextRunTime.Sub(time.Now()), func() {
        // Ensure that this task does not start too quickly.
        // We will acquire this lock for the duration of the task run later.
        taskLock.Lock();
        taskLock.Unlock();

        runTask(courseID, target, id, runFunc);

        // Schedule the next run.
        err := scheduleTask(courseID, target, id, runFunc, when);
        if (err != nil) {
            log.Error().Err(err).Str("task", target.String()).Str("when", when.String()).Msg("Failed to reschedule task.");
        }
    });

    _, ok := timers[courseID];
    if (!ok) {
        timers[courseID] = make(map[string]*timerInfo);
    }

    timers[courseID][id] = &timerInfo{
        ID: id,
        CourseID: courseID,
        Timer: timer,
        Lock: taskLock,
        Stopped: false,
    };

    log.Debug().Str("task", target.String()).Str("when", when.String()).Any("next-time", nextRunTime).Msg("Task scheduled.");

    return nil;
}

func stopCourseWithoutLocks(courseID string) {
    for _, timerInfo := range timers[courseID] {
        // Acquiring the lock ensure that we are not interrupting an active task.
        // If we get the lock before a scheduled task, then that task will never run.
        timerInfo.Lock.Lock();
        defer timerInfo.Lock.Unlock();

        timerInfo.Stopped = true;
        stoppedTasks[timerInfo.ID] = true;
        timerInfo.Timer.Stop();
    }

    delete(timers, courseID);
}

// Stop all the tasks associated with this course.
// This will block until all tasks have been stopped.
// Will wait for any already running tasks to finish.
func StopCourse(courseID string) {
    timersLock.Lock();
    defer timersLock.Unlock();

    stopCourseWithoutLocks(courseID);
}

// Stop all the tasks.
// This will block until all tasks have been stopped.
// Will wait for any already running tasks to finish.
func StopAll() {
    timersLock.Lock();
    defer timersLock.Unlock();

    keys := maps.Keys(timers);
    for _, key := range keys {
        stopCourseWithoutLocks(key);
    }
}

func runTask(courseID string, target tasks.ScheduledTask, id string, runFunc RunFunc) {
    info := getTimerInfo(courseID, id);
    if (info == nil) {
        return;
    }

    if (info.Stopped) {
        return;
    }

    info.Lock.Lock();
    defer info.Lock.Unlock();

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

// Should a task run?
func getTimerInfo(courseID string, id string) *timerInfo {
    timersLock.Lock();
    defer timersLock.Unlock();

    timers, ok := timers[courseID];
    if (!ok) {
        // No course timers are found, this course must have been stopped.
        return nil;
    }

    info, ok := timers[id];
    if (!ok) {
        // This timer cannot be found, it must have been stopped.
        return nil;
    }

    return info;
}