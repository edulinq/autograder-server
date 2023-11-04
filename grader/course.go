package grader

import (
    "fmt"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/model"
)

// TEST - This is going away.
var courses map[string]model.Course = make(map[string]model.Course);

func GetCourses() map[string]model.Course {
    return courses;
}

func ActivateCourses() error {
    for _, course := range courses {
        err := course.Activate();
        if (err != nil) {
            return fmt.Errorf("Unable to activate course '%s': '%w'.", course.GetID(), err);
        }
    }

    return nil;
}

func GetCourse(id string) model.Course {
    id, err := common.ValidateID(id);
    if (err != nil) {
        return nil;
    }

    course, ok := courses[id];
    if (!ok) {
        return nil;
    }

    return course;
}

func GetAssignment(courseID string, assignmentID string) model.Assignment {
    course := GetCourse(courseID);
    if (course == nil) {
        return nil;
    }

    assignmentID, err := common.ValidateID(assignmentID);
    if (err != nil) {
        return nil;
    }

    return course.GetAssignment(assignmentID);
}

// Get the course and assignment from identifiers.
func VerifyCourseAssignment(courseID string, assignmentID string) (model.Course, model.Assignment, error) {
    course := GetCourse(courseID);
    if (course == nil) {
        return nil, nil, fmt.Errorf("Unknown course: '%s'.", courseID);
    }

    assignment := GetAssignment(courseID, assignmentID);
    if (assignment == nil) {
        return nil, nil, fmt.Errorf("Unknown assignment: '%s'.", assignmentID);
    }

    return course, assignment, nil;
}
