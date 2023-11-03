package task

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model2"
    "github.com/eriq-augustine/autograder/scoring"
)

type ScoringUploadTask struct {
    Disable bool `json:"disable"`
    DryRun bool `json:"dry-run"`
    When ScheduledTime `json:"when"`

    course model2.Course `json:"-"`
}

func (this *ScoringUploadTask) Validate(course model2.Course) error {
    this.When.id = fmt.Sprintf("score-%s", course.GetID());

    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.Disable = (this.Disable || config.NO_TASKS.Get());

    this.course = course;

    if (this.course.GetLMSAdapter() == nil) {
        return fmt.Errorf("Score and Upload task course must have an LMS adapter.");
    }

    lmsIDs, assignmentIDs := this.course.GetAssignmentLMSIDs();
    for i, _ := range lmsIDs {
        if (lmsIDs[i] == "") {
            log.Warn().Str("course", course.GetID()).Str("assignment", assignmentIDs[i]).
                    Msg("Score and Upload course has an assignment with a missing LMS ID.");
        }
    }

    return nil;
}

func (this *ScoringUploadTask) String() string {
    return fmt.Sprintf("Score and Upload of course '%s' at '%s' (next time: '%s').", this.course.GetID(), this.When.String(), this.When.ComputeNext());
}

// Schedule this task to be regularly run at the scheduled time.
func (this *ScoringUploadTask) Schedule() {
    if (this.Disable) {
        return;
    }

    this.When.Schedule(func() {
        err := this.Run();
        if (err != nil) {
            log.Error().Err(err).Str("course", this.course.GetID()).Msg("Score and Upload task failed.");
        }
    });
}

// Stop any scheduled executions of this task.
func (this *ScoringUploadTask) Stop() {
    this.When.Stop();
}

// Run the task regardless of schedule.
func (this *ScoringUploadTask) Run() error {
    return scoring.FullCourseScoringAndUpload(this.course, this.DryRun);
}
