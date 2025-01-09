package scoring

import (
	"fmt"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

// Returns: {assignmentID: {email: scoringInfo, ...}, ...}.
func FullCourseScoringAndUpload(course *model.Course, dryRun bool) (map[string]map[string]*model.ScoringInfo, error) {
	assignments := course.GetSortedAssignments()

	log.Debug("Beginning full scoring for course.", course, log.NewAttr("dry-run", dryRun))

	results := make(map[string]map[string]*model.ScoringInfo, len(assignments))
	for i, assignment := range assignments {
		if assignment.GetLMSID() == "" {
			log.Warn("Assignment has no LMS id, skipping scoring.", course, assignment)
			continue
		}

		log.Trace("Scoring course assignment.", course, assignment,
			log.NewAttr("index", i),
			log.NewAttr("dry-run", dryRun))

		uploadedScores, err := FullAssignmentScoringAndUpload(assignment, dryRun)
		if err != nil {
			return nil, fmt.Errorf("Failed to grade assignment '%s' for course '%s': '%w'.", assignment.GetID(), course.GetID(), err)
		}

		results[assignment.GetID()] = uploadedScores
	}

	log.Debug("Finished full scoring for course.", course, log.NewAttr("dry-run", dryRun))

	return results, nil
}
