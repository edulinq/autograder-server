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

    CourseID string `json:"-"`
}

func (this *ScoringUploadTask) Validate(course Course) error {
    this.When.id = fmt.Sprintf("score-%s", course.GetID());
    this.CourseID = course.GetID();

    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.Disable = (this.Disable || config.NO_TASKS.Get());

    if (course.GetLMSAdapter() == nil) {
        return fmt.Errorf("Score and Upload task course must have an LMS adapter.");
    }

    lmsIDs, assignmentIDs := course.GetAssignmentLMSIDs();
    for i, _ := range lmsIDs {
        if (lmsIDs[i] == "") {
            log.Warn().Str("course", course.GetID()).Str("assignment", assignmentIDs[i]).
                    Msg("Score and Upload course has an assignment with a missing LMS ID.");
        }
    }

    return nil;
}

func (this *ScoringUploadTask) IsDisabled() bool {
    return this.Disable;
}

func (this *ScoringUploadTask) GetTime() *ScheduledTime {
    return &this.When;
}

func (this *ScoringUploadTask) String() string {
    return fmt.Sprintf("Score and Upload of course '%s': '%s'.",
            this.CourseID, this.When.String());
}
