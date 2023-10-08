package model

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
)

type ScoringUploadTask struct {
    Disable bool `json:"disable"`
    DryRun bool `json:"dry-run"`
    When ScheduledTime `json:"when"`

    course *Course `json:"-"`
}

func (this *ScoringUploadTask) Validate(course *Course) error {
    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.Disable = (this.Disable || config.NO_TASKS.GetBool());

    this.course = course;

    if (this.course.CanvasInstanceInfo == nil) {
        return fmt.Errorf("Score and Upload task course must have Canvas instance information.");
    }

    for _, assignment := range this.course.Assignments {
        if (assignment.CanvasID == "") {
            log.Warn().Str("course", course.ID).Str("assignment", assignment.ID).
                    Msg("Score and Upload assignment is missing a Canvas ID.");
        }
    }

    return nil;
}

func (this *ScoringUploadTask) String() string {
    return fmt.Sprintf("Score and Upload of course '%s' at '%s' (next time: '%s').", this.course.ID, this.When.String(), this.When.ComputeNext());
}

// Schedule this task to be regularly run at the scheduled time.
func (this *ScoringUploadTask) Schedule() {
    if (this.Disable) {
        return;
    }

    this.When.Schedule(func() {
        err := this.Run();
        if (err != nil) {
            log.Error().Err(err).Str("course", this.course.ID).Msg("Score and Upload task failed.");
        }
    });
}

// Stop any scheduled executions of this task.
func (this *ScoringUploadTask) Stop() {
    this.When.Stop();
}

// Run the task regardless of schedule.
func (this *ScoringUploadTask) Run() error {
    return this.course.FullScoringAndUpload(this.DryRun);
}
