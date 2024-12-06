package model

import (
	"fmt"

	"github.com/edulinq/autograder/internal/util"
)

const ASSIGNMENT_CONFIG_FILENAME = "assignment.json"

// Load an assignment config from a given path.
// If this assignment is being loaded from a source dir, then give the relative path to the assignment's dir
// relative to the course's base dir.
func ReadAssignmentConfig(course *Course, path string, relSourceDir string) (*Assignment, error) {
	if course == nil {
		return nil, fmt.Errorf("Cannot load an assignment without a course.")
	}

	if !util.IsFile(path) {
		return nil, fmt.Errorf("Assignment path does not exist or is not a file: '%s'.", path)
	}

	var assignment Assignment
	err := util.JSONFromFile(path, &assignment)
	if err != nil {
		return nil, fmt.Errorf("Could not load assignment config (%s): '%w'.", path, err)
	}

	assignment.Course = course

	// Only override the relative source dir if it is empty.
	// This should only be the first time that we load the course from source (not from the DB).
	if (assignment.RelSourceDir == "") && (relSourceDir != "") {
		assignment.RelSourceDir = relSourceDir
	}

	err = assignment.Validate()
	if err != nil {
		return nil, fmt.Errorf("Failed to validate assignment config (%s): '%w'.", path, err)
	}

	err = course.AddAssignment(&assignment)
	if err != nil {
		return nil, fmt.Errorf("Failed to add assignment to course (%s): '%w'.", path, err)
	}

	return &assignment, nil
}
