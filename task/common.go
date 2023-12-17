// Implementations for running tasks.
package task

import (
    "fmt"
    "time"
    "sync"

    "github.com/rs/zerolog/log"
    "golang.org/x/exp/maps"

    "github.com/eriq-augustine/autograder/common"
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
// The key (timerID) is the same as timerInfo.ID.
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

    // Does this task need to be run right now
    // to cachup for lost time (e.g. the server being down).
    catchup, err := checkForCatchup(course.GetID(), target);
    if (err != nil) {
        return err;
    }

    if (catchup) {
        // Special ID for catchup tasks.
        timerID := fmt.Sprintf("%s::catchup", target.GetID());

        err := scheduleTask(course.GetID(), target, timerID, runFunc, nil);
        if (err != nil) {
            return fmt.Errorf("Failed to schedule catchup task (%s): '%w'.", target.GetID(), err);
        }
    }

    for i, when := range target.GetTimes() {
        // ID unique to every (task, timer).
        timerID := fmt.Sprintf("%s::%03d", target.GetID(), i);

        timersLock.Lock();
        delete(stoppedTasks, timerID);
        timersLock.Unlock();

        err := scheduleTask(course.GetID(), target, timerID, runFunc, when);
        if (err != nil) {
            return fmt.Errorf("Failed to schedule task (%s): '%w'.", target.GetID(), err);
        }
    }

    return nil;
}

// Check to see if it has been too long since this task has been run.
// Do this by getting the minimum duration for all the task's timers,
// and seeing if it has been at least that long since the task has been run.
// Return true of a catchup task needs to be run.
func checkForCatchup(courseID string, target tasks.ScheduledTask) (bool, error) {
    minDuration, exists := target.GetMinDuration();
    if (!exists) {
        return false, nil;
    }

    lastRunTime, err := db.GetLastTaskCompletion(courseID, target.GetID());
    if (err != nil) {
        return false, fmt.Errorf("Failed to get last time task was run: '%w'.", err);
    }

    // Don't catchup if the task has never been run.
    if (lastRunTime.IsZero()) {
        return false, nil;
    }

    currentDuration := time.Now().Sub(lastRunTime);
    return (currentDuration > minDuration), nil;
}

// Schedule a task.
// |when| will be nil on a catchup task.
func scheduleTask(courseID string, target tasks.ScheduledTask, timerID string, runFunc RunFunc, when *common.ScheduledTime) error {
    timersLock.Lock();
    defer timersLock.Unlock();

    // Ensure this task has not been stopped,
    // and therefore should not be scheduled.
    stopped, exists := stoppedTasks[timerID];
    if (exists && stopped) {
        return nil;
    }

    // In addition to the general timers lock, create a lock just for this timer that prevents
    // a scheduled task from calling too soon (before the timerInfo is setup).
    taskLock := &sync.Mutex{};
    taskLock.Lock();
    defer taskLock.Unlock();

    nextRunDuration := 5 * time.Microsecond;
    if (when != nil) {
        nextRunTime := when.ComputeNextTimeFromNow();
        nextRunDuration = nextRunTime.Sub(time.Now());
    }

    timer := time.AfterFunc(nextRunDuration, func() {
        // Ensure that this task does not start too quickly.
        // We will acquire this lock for the duration of the task run later.
        taskLock.Lock();
        taskLock.Unlock();

        runTask(courseID, target, timerID, runFunc);

        // Do not reschedule catchup tasks.
        if (when == nil) {
            return;
        }

        // Schedule the next run.
        err := scheduleTask(courseID, target, timerID, runFunc, when);
        if (err != nil) {
            log.Error().Err(err).Str("task", target.GetID()).Str("when", when.String()).Msg("Failed to reschedule task.");
        }
    });

    _, ok := timers[courseID];
    if (!ok) {
        timers[courseID] = make(map[string]*timerInfo);
    }

    timers[courseID][timerID] = &timerInfo{
        ID: timerID,
        CourseID: courseID,
        Timer: timer,
        Lock: taskLock,
        Stopped: false,
    };

    if (when == nil) {
        log.Debug().Str("task", target.GetID()).Msg("Catchup task scheduled.");
    } else {
        nextRunTime := time.Now().Add(nextRunDuration);
        log.Debug().Str("task", target.GetID()).Str("when", when.String()).Any("next-time", nextRunTime).Msg("Task scheduled.");
    }

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

func runTask(courseID string, target tasks.ScheduledTask, timerID string, runFunc RunFunc) {
    target.GetLock().Lock();
    defer target.GetLock().Unlock();

    taskID := target.GetID();
    now := time.Now();

    info := getTimerInfo(courseID, timerID);
    if (info == nil) {
        return;
    }

    if (info.Stopped) {
        return;
    }

    info.Lock.Lock();
    defer info.Lock.Unlock();

    lastRunTime, err := db.GetLastTaskCompletion(courseID, taskID);
    if (err != nil) {
        log.Error().Err(err).Str("course-id", courseID).Str("task", taskID).Msg("Failed to get last time task was run.");
        // Keep trying to run the task.
    }

    // Skip this task if it was run too recently.
    lastRunDuration := now.Sub(lastRunTime);
    if (lastRunDuration < (time.Duration(config.TASK_MIN_REST_SECS.Get()) * time.Second)) {
        log.Debug().Str("course-id", courseID).Str("task", taskID).Any("last-run", lastRunTime).
                Msg("Skipping task run, last run was too recent.");
        return;
    }

    log.Debug().Str("task", taskID).Msg("Task started.");

    course, err := db.GetCourse(courseID);
    if (err != nil) {
        log.Error().Err(err).Str("course-id", courseID).Str("task", taskID).Msg("Failed to get course for task.");
        return;
    }

    if (course == nil) {
        log.Error().Str("course-id", courseID).Str("task", taskID).Msg("Could not find course for task.");
        return;
    }

    err = runFunc(course, target);
    if (err != nil) {
        log.Error().Err(err).Str("course-id", courseID).Str("task", taskID).Msg("Task run failed.");
        return;
    }

    log.Debug().Str("course-id", courseID).Str("task", taskID).Msg("Task finished.");

    err = db.LogTaskCompletion(courseID, taskID, now);
    if (err != nil) {
        log.Error().Err(err).Str("course-id", courseID).Str("task", taskID).Msg("Failed to log task completion.");
        return;
    }
}

// Should a task run?
func getTimerInfo(courseID string, timerID string) *timerInfo {
    timersLock.Lock();
    defer timersLock.Unlock();

    timers, ok := timers[courseID];
    if (!ok) {
        // No course timers are found, this course must have been stopped.
        return nil;
    }

    info, ok := timers[timerID];
    if (!ok) {
        // This timer cannot be found, it must have been stopped.
        return nil;
    }

    return info;
}
