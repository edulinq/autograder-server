package db

import (
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/model"
)

// Get an assignment or panic.
// This is a convenience function for the CLI mains that need to get an assignment.
func MustGetAssignment(rawCourseID string, rawAssignmentID string) model.Assignment {
    course := MustGetCourse(rawCourseID);

    assignmentID, err := common.ValidateID(rawAssignmentID);
    if (err != nil) {
        log.Fatal().Err(err).Str("assignment-id", rawAssignmentID).Msg("Failed to validate assignment id.");
    }

    assignment := course.GetAssignment(assignmentID);
    if (assignment == nil) {
        log.Fatal().Str("course-id", course.GetID()).Str("assignment-id", assignmentID).Msg("Could not find assignment.");
    }

    return assignment;
}
