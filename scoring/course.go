package scoring

import (
    "fmt"

    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/model"
)

func FullCourseScoringAndUpload(course *model.Course, dryRun bool) error {
    assignments := course.GetSortedAssignments();

    log.Debug("Beginning full scoring for course.",
            log.NewAttr("course", course.GetID()),
            log.NewAttr("dry-run", dryRun));

    for i, assignment := range assignments {
        if (assignment.GetLMSID() == "") {
            log.Warn("Assignment has no LMS id, skipping scoring.",
                    log.NewAttr("course", course.GetID()),
                    log.NewAttr("assignment", assignment.GetID()));
            continue;
        }

        log.Debug("Scoring course assignment.",
                log.NewAttr("course", course.GetID()),
                log.NewAttr("assignment", assignment.GetID()),
                log.NewAttr("index", i),
                log.NewAttr("dry-run", dryRun));

        err := FullAssignmentScoringAndUpload(assignment, dryRun);
        if (err != nil) {
            return fmt.Errorf("Failed to grade assignment '%s' for course '%s': '%w'.", course.GetID(), assignment.GetID(), err);
        }
    }

    log.Debug("Finished full scoring for course.",
            log.NewAttr("course", course.GetID()), log.NewAttr("dry-run", dryRun));

    return nil;
}
