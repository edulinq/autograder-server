package tasks

import (
    "fmt"

    "github.com/eriq-augustine/autograder/log"
)

type ScoringUploadTask struct {
    *BaseTask

    DryRun bool `json:"dry-run"`
}

func (this *ScoringUploadTask) Validate(course TaskCourse) error {
    this.BaseTask.Name = "scoring";

    err := this.BaseTask.Validate(course);
    if (err != nil) {
        return err;
    }

    if (!course.HasLMSAdapter()) {
        return fmt.Errorf("Score and Upload task course must have an LMS adapter.");
    }

    lmsIDs, assignmentIDs := course.GetAssignmentLMSIDs();
    for i, _ := range lmsIDs {
        if (lmsIDs[i] == "") {
            log.Warn("Score and Upload course has an assignment with a missing LMS ID.",
                    log.NewCourseAttr(course.GetID()), log.NewAssignmentAttr(assignmentIDs[i]));
        }
    }

    return nil;
}
