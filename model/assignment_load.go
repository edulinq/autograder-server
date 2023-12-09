package model

import (
    "fmt"
    "path/filepath"

    "github.com/eriq-augustine/autograder/util"
)

const ASSIGNMENT_CONFIG_FILENAME = "assignment.json"

// Load an assignment config from a given JSON path.
func LoadAssignment(course *Course, path string) (*Assignment, error) {
    if (course == nil) {
        return nil, fmt.Errorf("Cannot load an assignment without a course.");
    }

    var assignment Assignment;
    err := util.JSONFromFile(path, &assignment);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load assignment config (%s): '%w'.", path, err);
    }

    assignment.Course = course;

    if (assignment.SourceDir == "") {
        // Force the source dir to be relative to the course dir.
        assignment.SourceDir, err = filepath.Rel(course.GetSourceDir(), util.ShouldAbs(filepath.Dir(path)));
        if (err != nil) {
            return nil, fmt.Errorf("Could not compute source dir for assignment (%s): '%w'.", path, err);
        }
    }

    err = assignment.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate assignment config (%s): '%w'.", path, err);
    }

    otherAssignment := course.GetAssignment(assignment.GetID());
    if (otherAssignment != nil) {
        return nil, fmt.Errorf(
                "Found multiple assignments with the same ID ('%s'): ['%s', '%s'].",
                assignment.GetID(), otherAssignment.GetSourceDir(), assignment.GetSourceDir());
    }
    course.Assignments[assignment.GetID()] = &assignment;

    return &assignment, nil;
}
