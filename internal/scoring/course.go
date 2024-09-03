package scoring

import (
	"fmt"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

func FullCourseScoringAndUpload(course *model.Course, dryRun bool) error {
	assignments := course.GetSortedAssignments()

	log.Debug("Beginning full scoring for course.", course, log.NewAttr("dry-run", dryRun))

	for i, assignment := range assignments {
		if assignment.GetLMSID() == "" {
			log.Warn("Assignment has no LMS id, skipping scoring.", course, assignment)
			continue
		}

		log.Trace("Scoring course assignment.", course, assignment,
			log.NewAttr("index", i),
			log.NewAttr("dry-run", dryRun))

		err := FullAssignmentScoringAndUpload(assignment, dryRun)
		if err != nil {
			return fmt.Errorf("Failed to grade assignment '%s' for course '%s': '%w'.", assignment.GetID(), course.GetID(), err)
		}
	}

	log.Debug("Finished full scoring for course.", course, log.NewAttr("dry-run", dryRun))

	return nil
}
