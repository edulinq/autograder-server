package grader

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const COURSE_CONFIG_FILENAME = "course.json"

// Discover all courses (from the config) and load all the associated courses and assignments.
func LoadCourses() (map[string]*model.Course, error) {
    baseDir := config.GetString(config.COURSES_ROOTDIR);

    configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err);
    }

    log.Info().Int("count", len(configPaths)).Msg(fmt.Sprintf("Found %d course config(s).", len(configPaths)));

    courses := make(map[string]*model.Course);

    for _, configPath := range configPaths {
        courseConfig, err := model.LoadCourseDirectory(configPath);
        if (err != nil) {
            return nil, fmt.Errorf("Could not load course config '%s': '%w'.", configPath, err);
        }

        courses[courseConfig.ID] = courseConfig;

        log.Info().Str("path", configPath).Str("id", courseConfig.ID).Int("assignments", len(courseConfig.Assignments)).Msg("Loading course config.");
    }

    return courses, nil;
}
