package db

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db/types"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// Get a course.
func GetCourse(rawCourseID string) (model.Course, error) {
    courseID, err := common.ValidateID(rawCourseID);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate course id '%s': '%w'.", rawCourseID, err);
    }

    return backend.GetCourse(courseID);
}

// Get a course or panic.
// This is a convenience function for the CLI mains that need to get a course.
func MustGetCourse(rawCourseID string) model.Course {
    course, err := GetCourse(rawCourseID);
    if (err != nil) {
        log.Fatal().Err(err).Str("course-id", rawCourseID).Msg("Failed to get course.");
    }

    if (course == nil) {
        log.Fatal().Str("course-id", rawCourseID).Msg("Could not find course.");
    }

    return course;
}

// Get all the known courses.
func GetCourses() (map[string]model.Course, error) {
    dbCourses, err := backend.GetCourses();
    if (err != nil) {
        return nil, err;
    }

    courses := make(map[string]model.Course, len(dbCourses));
    for key, value := range dbCourses {
        courses[key] = value;
    }

    return courses, nil;
}

// Get all the known courses or panic.
// This is a convenience function for the CLI mains.
func MustGetCourses() map[string]model.Course {
    courses, err := GetCourses();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to get courses.");
    }

    return courses;
}

func LoadCourse(path string) error {
    return backend.LoadCourse(path);
}

// TEST - Do we need this anymore?
func MustLoadCourse(path string) {
    err := LoadCourse(path);
    if (err != nil) {
        log.Fatal().Err(err).Str("path", path).Msg("Failed to load course.");
    }
}

func SaveCourse(rawCourse model.Course) error {
    course, ok := rawCourse.(*types.Course);
    if (!ok) {
        return fmt.Errorf("Course '%v' is not a db course.", rawCourse);
    }

    return backend.SaveCourse(course);
}

// Search the courses root directory and load all the associated courses and assignments.
func LoadCourses() error {
    return LoadCoursesFromDir(config.COURSES_ROOT.Get());
}

func LoadCoursesFromDir(baseDir string) error {
    log.Debug().Str("dir", baseDir).Msg("Searching for courses.");

    configPaths, err := util.FindFiles(types.COURSE_CONFIG_FILENAME, baseDir);
    if (err != nil) {
        return fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err);
    }

    log.Info().Int("count", len(configPaths)).Msg(fmt.Sprintf("Found %d course config(s).", len(configPaths)));

    for _, configPath := range configPaths {
        err := LoadCourse(configPath);
        if (err != nil) {
            return fmt.Errorf("Could not load course '%s': '%w'.", configPath, err);
        }

        log.Debug().Str("path", configPath).Msg("Loaded course.");
    }

    return nil;
}
