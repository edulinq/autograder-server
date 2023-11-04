package types

import (
    "fmt"
    "path/filepath"

    "github.com/eriq-augustine/autograder/util"
)

const COURSE_CONFIG_FILENAME = "course.json"

func LoadCourse(path string) (*Course, error) {
    course, err := loadCourseConfig(path);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config at '%s': '%w'.", path, err);
    }

    courseDir := filepath.Dir(path);

    assignmentPaths, err := util.FindFiles(ASSIGNMENT_CONFIG_FILENAME, courseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for assignment configs in '%s': '%w'.", courseDir, err);
    }

    for _, assignmentPath := range assignmentPaths {
        _, err := LoadAssignment(assignmentPath, course);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to load assignment config '%s': '%w'.", assignmentPath, err);
        }
    }

    return course, nil;
}

func loadCourseConfig(path string) (*Course, error) {
    var course Course;
    err := util.JSONFromFile(path, &course);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config (%s): '%w'.", path, err);
    }

    course.SourcePath = util.ShouldAbs(path);
    course.Assignments = make(map[string]*Assignment);

    err = course.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Could not validate course config (%s): '%w'.", path, err);
    }

    return &course, nil;
}

// Check this directory and all parent directories for a course config file.
func loadParentCourse(basepath string) (*Course, error) {
    configPath := util.SearchParents(basepath, COURSE_CONFIG_FILENAME);
    if (configPath == "") {
        return nil, fmt.Errorf("Could not locate course config.");
    }

    return LoadCourse(configPath);
}
