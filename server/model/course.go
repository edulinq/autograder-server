package model

import (
	"fmt"
    "path/filepath"

    "github.com/eriq-augustine/autograder/util"
)

const COURSE_CONFIG_FILENAME = "course.json"

type Course struct {
    ID string  `json:"id"`
    DisplayName string `json:"display-name"`

    // Ignore these fields in JSON.
    Assignments map[string]*Assignment `json:"-"`
}

func LoadCourseConfig(path string) (*Course, error) {
    var config Course;
    err := util.JSONFromFile(path, &config);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config (%s): '%w'.", path, err);
    }

    config.Assignments = make(map[string]*Assignment);

    err = config.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Could not validate course config (%s): '%w'.", path, err);
    }

    return &config, nil;
}

// Load the course (with its JSON config) and all assignments (JSON configs) recursivley in a directory.
// The path should point to the course config,
// and the directory that path lives in will be searched for assignment configs.
func LoadCourseDirectory(courseConfigPath string) (*Course, error) {
    courseConfig, err := LoadCourseConfig(courseConfigPath);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config at '%s': '%w'.", courseConfigPath, err);
    }

    courseDir := filepath.Dir(courseConfigPath);

    assignmentPaths, err := util.FindFiles(ASSIGNMENT_CONFIG_FILENAME, courseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for assignment configs in '%s': '%w'.", courseDir, err);
    }

    for _, assignmentPath := range assignmentPaths {
        _, err := LoadAssignmentConfig(assignmentPath, courseConfig);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to load assignment config '%s': '%w'.", assignmentPath, err);
        }
    }

    return courseConfig, nil;
}

func (this *Course) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = ValidateID(this.ID);
    if (err != nil) {
        return err;
    }

    return nil;
}

// Ensure the course is ready for grading.
func (this *Course) Init() error {
    for _, assignment := range this.Assignments {
        err := assignment.Init()
        if (err != nil) {
            return err;
        }
    }

    return nil;
}

// Check this directory and all parent directories for a course config file.
func loadParentCourseConfig(basepath string) (*Course, error) {
    basepath = util.MustAbs(basepath);

    for ; ; {
        configPath := filepath.Join(basepath, COURSE_CONFIG_FILENAME);

        if (!util.PathExists(configPath)) {
            // Move up to the parent.
            oldBasepath := basepath;
            basepath = filepath.Dir(basepath);

            if (oldBasepath == basepath) {
                break;
            }

            continue;
        }

        return LoadCourseConfig(configPath);
    }

    return nil, fmt.Errorf("Could not locate course config.");
}
