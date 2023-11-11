package db

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/model"
)

// Get an assignment or panic.
// This is a convenience function for the CLI mains that need to get an assignment.
func MustGetAssignment(rawCourseID string, rawAssignmentID string) model.Assignment {
    assignment, err := GetAssignment(rawCourseID, rawAssignmentID);
    if (err != nil) {
        log.Fatal().Err(err).
                Str("raw-course-id", rawCourseID).Str("raw-assignment-id", rawAssignmentID).
                Msg("Failed to get assignment.");
    }

    return assignment;
}

func GetAssignment(rawCourseID string, rawAssignmentID string) (model.Assignment, error) {
    course, err := GetCourse(rawCourseID);
    if (err != nil) {
        return nil, err;
    }

    assignmentID, err := common.ValidateID(rawAssignmentID);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate assignment id '%s': '%w'.", rawAssignmentID, err);
    }

    assignment := course.GetAssignment(assignmentID);
    if (assignment == nil) {
        return nil, fmt.Errorf("Unable to find assignment '%s' for course '%s'.", assignmentID, course.GetID());
    }

    return assignment, nil;
}
