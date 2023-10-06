package grader

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var courses map[string]*model.Course = make(map[string]*model.Course);

func GetCourses() map[string]*model.Course {
    return courses;
}

// Discover all courses (from the config) and load all the associated courses and assignments.
func LoadCourses() error {
    return LoadCoursesFromDir(config.COURSES_ROOT.GetString());
}

func LoadCoursesFromDir(baseDir string) error {
    log.Debug().Str("dir", baseDir).Msg("Searching for courses.");

    configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir);
    if (err != nil) {
        return fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err);
    }

    log.Info().Int("count", len(configPaths)).Msg(fmt.Sprintf("Found %d course config(s).", len(configPaths)));

    for _, configPath := range configPaths {
        courseConfig, err := model.LoadCourseDirectory(configPath);
        if (err != nil) {
            return fmt.Errorf("Could not load course config '%s': '%w'.", configPath, err);
        }

        courses[courseConfig.ID] = courseConfig;

        log.Info().Str("path", configPath).Str("id", courseConfig.ID).Int("assignments", len(courseConfig.Assignments)).Msg("Loading course config.");
    }

    return nil;
}

func GetCourse(id string) *model.Course {
    id, err := model.ValidateID(id);
    if (err != nil) {
        return nil;
    }

    course, ok := courses[id];
    if (!ok) {
        return nil;
    }

    return course;
}

func GetAssignment(courseID string, assignmentID string) *model.Assignment {
    course := GetCourse(courseID);
    if (course == nil) {
        return nil;
    }

    assignmentID, err := model.ValidateID(assignmentID);
    if (err != nil) {
        return nil;
    }

    assignment, ok := course.Assignments[assignmentID];
    if (!ok) {
        return nil;
    }

    return assignment;
}

// Get the course and assignment from identifiers.
func VerifyCourseAssignment(courseID string, assignmentID string) (*model.Course, *model.Assignment, error) {
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
