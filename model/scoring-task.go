package model

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
)

type ScoringUploadTask struct {
    Disable bool `json:"disable"`
    When ScheduledTime `json:"when"`

    assignment *Assignment `json:"-"`
}

func (this *ScoringUploadTask) Validate(assignment *Assignment) error {
    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.Disable = (this.Disable || config.NO_TASKS.GetBool());

    this.assignment = assignment;

    if (this.assignment.CanvasID == "") {
        return fmt.Errorf("Score and Upload task assignment must have a Canvas ID.");
    }

    return nil;
}

func (this *ScoringUploadTask) String() string {
    return fmt.Sprintf("Score and Upload of assignmnet '%s' at '%s' (next time: '%s').", this.assignment.ID, this.When.String(), this.When.ComputeNext());
}

// Schedule this task to be regularly run at the scheduled time.
func (this *ScoringUploadTask) Schedule() {
    if (this.Disable) {
        return;
    }

    this.When.Schedule(func() {
        err := this.Run();
        if (err != nil) {
            log.Error().Err(err).Str("assignment", this.assignment.ID).Msg("Score and Upload task failed.");
        }
    });
}

// Stop any scheduled executions of this task.
func (this *ScoringUploadTask) Stop() {
    this.When.Stop();
}

// Run the task regardless of schedule.
func (this *ScoringUploadTask) Run() error {
    return this.assignment.FullScoringAndUpload(false);
}
