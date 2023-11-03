package scoring

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/model2"
)

func FullCourseScoringAndUpload(course model2.Course, dryRun bool) error {
    assignments := course.GetSortedAssignments();

    log.Debug().Str("course", course.GetID()).Bool("dry-run", dryRun).Msg("Beginning full scoring for course.");

    for i, assignment := range assignments {
        if (assignment.GetLMSID() == "") {
            log.Warn().Str("course", course.GetID()).Str("assignment", assignment.GetID()).Msg("Assignment has no LMS id, skipping scoring.");
            continue;
        }

        log.Debug().Str("course", course.GetID()).Str("assignment", assignment.GetID()).Int("index", i).Bool("dry-run", dryRun).
                Msg("Scoring course assignment.");

        err := FullAssignmentScoringAndUpload(assignment, dryRun);
        if (err != nil) {
            return fmt.Errorf("Failed to grade assignment '%s' for course '%s': '%w'.", course.GetID(), assignment.GetID(), err);
        }
    }

    log.Debug().Str("course", course.GetID()).Bool("dry-run", dryRun).Msg("Finished full scoring for course.");

    return nil;
}
