package grader

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const COURSE_CONFIG_FILENAME = "course.json"

var courses map[string]*model.Course = make(map[string]*model.Course);

func GetCourses() map[string]*model.Course {
    return courses;
}

// Discover all courses (from the config) and load all the associated courses and assignments.
func LoadCourses() error {
    baseDir := config.GetString(config.COURSES_ROOTDIR);

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

    assignment, ok := course.Assignments[assignmentID];
    if (!ok) {
        return nil;
    }

    return assignment;
}
