package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

// Get an assignment or panic.
// This is a convenience function for the CLI mains that need to get an assignment.
func MustGetAssignment(rawCourseID string, rawAssignmentID string) *model.Assignment {
	assignment, err := GetAssignment(rawCourseID, rawAssignmentID)
	if err != nil {
		log.Fatal("Failed to get assignment.",
			err, log.NewCourseAttr(rawCourseID), log.NewAssignmentAttr(rawAssignmentID))
	}

	return assignment
}

func GetAssignment(rawCourseID string, rawAssignmentID string) (*model.Assignment, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	course, err := GetCourse(rawCourseID)
	if err != nil {
		return nil, err
	}

	assignmentID, err := common.ValidateID(rawAssignmentID)
	if err != nil {
		return nil, fmt.Errorf("Failed to validate assignment id '%s': '%w'.", rawAssignmentID, err)
	}

	assignment := course.GetAssignment(assignmentID)
	if assignment == nil {
		return nil, fmt.Errorf("Unable to find assignment '%s' for course '%s'.", assignmentID, course.GetID())
	}

	return assignment, nil
}

func MustSaveAssignment(assignment *model.Assignment) {
	err := SaveAssignment(assignment)
	if err != nil {
		log.Fatal("Failed to save course.", err)
	}
}

func SaveAssignment(assignment *model.Assignment) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	err := assignment.Validate()
	if err != nil {
		return fmt.Errorf("Assignment '%s' is not valid: '%w'.", assignment.GetID(), err)
	}

	return backend.SaveAssignment(assignment)
}
