// Implementations for running tasks.
package task

import (
    "time"

    // TEST
    // "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
)

// TEST - Backup tasks need to remain accessible, so they can be stopped.
// TEST - They should probably not hold their own schedulers.
// TEST - Tasks should really just schedule and only be handed a course once it is ready to run.
var timers map[string]*time.Timer;

type RunFunc func(model.Course, model.ScheduledTask) error;

// TEST -- Just stub these out for now.
func Schedule(course model.Course, target model.ScheduledTask) (string, error) {
    return "", nil;
}

func Stop(id string) error {
    return nil;
}

func StopAll() error {
    return nil;
}

/* TEST

// Schedule this task to be regularly run at the scheduled time.
// Return an id that can be used to stop the task.
func Schedule(course model.Course, target model.ScheduledTask) (string, error) {
    if (target.IsDisabled() || config.NO_TASKS.Get()) {
        return "", nil;
    }

    var runFunc RunFunc;
    switch (target.(type)) {
        case *model.BackupTask:
            runFunc = RunBackupTask;
        default:
            return "", fmt.Errorf("Unknown task type: %t (%v).", target, target);
    }

    this.When.Schedule(func() {
        err := this.Run();
        if (err != nil) {
            log.Error().Err(err).Str("source", this.source).Str("dest", this.dest).Msg("Backup task failed.");
        }
    });

    nextRunTime := target.GetTime().ComputeNext();
    timer := time.AfterFunc(nextRunTime.Sub(time.Now()), func() {
        runTask(course.GetID(), target, runFunc);
    });

    log.Debug().Str("id", this.id).Any("next-time", this.nextRun).Msg("Task scheduled.");
}

// Stop any scheduled executions of this task.
func Stop(id string) {
    this.When.Stop();
}

// Run the task regardless of schedule.
func (this *BackupTask) Run() error {
    return RunBackup(this.source, this.dest, this.basename);
}

// TEST

// Set a recurring invoation of the given task.
// The caller is responsible for stopping any previous timers
// (or ignoring them if they no longer need to be canceled).
func (this *ScheduledTime) Schedule(task func()) {
    this.nextRun = this.ComputeNext();
    this.timer = time.AfterFunc(this.nextRun.Sub(time.Now()), func() {
        task()
        this.Schedule(task);
    });

    log.Debug().Str("id", this.id).Any("next-time", this.nextRun).Msg("Task scheduled.");
}

func (this *ScheduledTime) Stop() {
    if (this.timer == nil) {
        return;
    }

    this.nextRun = time.Time{}

    if (!this.timer.Stop()) {
        // Clear the channel.
        <- this.timer.C;
    }
    this.timer = nil;
}
*/
