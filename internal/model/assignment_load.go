package model

import (
    "fmt"
    "path/filepath"

    "github.com/edulinq/autograder/internal/util"
)

const ASSIGNMENT_CONFIG_FILENAME = "assignment.json"

// Load an assignment config from a given JSON path.
func ReadAssignmentConfig(course *Course, path string) (*Assignment, error) {
    if (course == nil) {
        return nil, fmt.Errorf("Cannot load an assignment without a course.");
    }

    var assignment Assignment;
    err := util.JSONFromFile(path, &assignment);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load assignment config (%s): '%w'.", path, err);
    }

    assignment.Course = course;

    if (assignment.RelSourceDir == "") {
        // Force the source dir to be relative to the course's base source dir.
        assignment.RelSourceDir, err = filepath.Rel(course.GetBaseSourceDir(), util.ShouldAbs(filepath.Dir(path)));
        if (err != nil) {
            return nil, fmt.Errorf("Could not compute relative source dir for assignment (%s): '%w'.", path, err);
        }
    }

    err = assignment.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate assignment config (%s): '%w'.", path, err);
    }

    err = course.AddAssignment(&assignment);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to add assignment to course (%s): '%w'.", path, err);
    }

    return &assignment, nil;
}
