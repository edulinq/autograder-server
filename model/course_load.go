package model

import (
    "fmt"
    "path/filepath"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

const COURSE_CONFIG_FILENAME = "course.json"

func LoadCourse(path string) (*Course, error) {
    course, err := LoadCourseConfig(path);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config at '%s': '%w'.", path, err);
    }

    courseDir := filepath.Dir(path);

    assignmentPaths, err := util.FindFiles(ASSIGNMENT_CONFIG_FILENAME, courseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for assignment configs in '%s': '%w'.", courseDir, err);
    }

    for _, assignmentPath := range assignmentPaths {
        _, err := LoadAssignment(course, assignmentPath);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to load assignment config '%s': '%w'.", assignmentPath, err);
        }
    }

    return course, nil;
}

func FullLoadCourse(path string) (*Course, map[string]*usr.User, []*artifact.GradingResult, error) {
    course, err := LoadCourse(path);
    if (err != nil) {
        return nil, nil, nil, err;
    }

    users, err := loadStaticUsers(path);
    if (err != nil) {
        return nil, nil, nil, fmt.Errorf("Failed to load static users for course config '%s': '%w'.", path, err);
    }

    submissions, err := loadStaticSubmissions(path);
    if (err != nil) {
        return nil, nil, nil, fmt.Errorf("Failed to load static submissions for course config '%s': '%w'.", path, err);
    }

    return course, users, submissions, nil;
}

func LoadCourseConfig(path string) (*Course, error) {
    var course Course;
    err := util.JSONFromFile(path, &course);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config (%s): '%w'.", path, err);
    }

    if (course.SourceDir == "") {
        course.SourceDir = util.ShouldAbs(filepath.Dir(path));
    }

    course.Assignments = make(map[string]*Assignment);

    err = course.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Could not validate course config (%s): '%w'.", path, err);
    }

    return &course, nil;
}
