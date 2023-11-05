package types

import (
    "fmt"
    "path/filepath"

    "github.com/eriq-augustine/autograder/util"
)

const ASSIGNMENT_CONFIG_FILENAME = "assignment.json"

// Load an assignment config from a given JSON path.
// If the course is nil, search all parent directories for the course config.
func LoadAssignment(path string, course *Course) (*Assignment, error) {
    var assignment Assignment;
    err := util.JSONFromFile(path, &assignment);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load assignment config (%s): '%w'.", path, err);
    }

    assignment.SourcePath = util.ShouldAbs(path);

    if (course == nil) {
        course, err = loadParentCourse(filepath.Dir(path));
        if (err != nil) {
            return nil, fmt.Errorf("Could not load course config for '%s': '%w'.", path, err);
        }
    }
    assignment.Course = course;

    err = assignment.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate assignment config (%s): '%w'.", path, err);
    }

    otherAssignment, ok := course.GetAssignment(assignment.GetID());
    if (ok) {
        return nil, fmt.Errorf(
                "Found multiple assignments with the same ID ('%s'): ['%s', '%s'].",
                assignment.GetID(), otherAssignment.GetSourceDir(), assignment.GetSourceDir());
    }
    course.Assignments[assignment.GetID()] = &assignment;

    return &assignment, nil;
}
