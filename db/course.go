package db

import (
    "fmt"
    "path/filepath"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// Get a course.
func GetCourse(rawCourseID string) (*model.Course, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    courseID, err := common.ValidateID(rawCourseID);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate course id '%s': '%w'.", rawCourseID, err);
    }

    course, err := backend.GetCourse(courseID);
    if (err != nil) {
        return nil, err;
    }

    if (course == nil) {
        return nil, nil;
    }

    return course, nil;
}

// Get a course or panic.
// This is a convenience function for the CLI mains that need to get a course.
func MustGetCourse(rawCourseID string) *model.Course {
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
func GetCourses() (map[string]*model.Course, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    dbCourses, err := backend.GetCourses();
    if (err != nil) {
        return nil, err;
    }

    courses := make(map[string]*model.Course, len(dbCourses));
    for key, value := range dbCourses {
        courses[key] = value;
    }

    return courses, nil;
}

// Get all the known courses or panic.
// This is a convenience function for the CLI mains.
func MustGetCourses() map[string]*model.Course {
    courses, err := GetCourses();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to get courses.");
    }

    return courses;
}

func ReloadCourse(course *model.Course) (string, error) {
    if (backend == nil) {
        return "", fmt.Errorf("Database has not been opened.");
    }

    path := filepath.Join(course.GetSourceDir(), model.COURSE_CONFIG_FILENAME);
    return backend.LoadCourse(path);
}

func LoadCourse(path string) (string, error) {
    if (backend == nil) {
        return "", fmt.Errorf("Database has not been opened.");
    }

    return backend.LoadCourse(path);
}

func MustLoadCourse(path string) string {
    courseID, err := LoadCourse(path);
    if (err != nil) {
        log.Fatal().Err(err).Str("path", path).Msg("Failed to load course.");
    }

    return courseID;
}

func SaveCourse(course *model.Course) error {
    if (backend == nil) {
        return fmt.Errorf("Database has not been opened.");
    }

    return backend.SaveCourse(course);
}

func DumpCourse(course *model.Course, targetDir string) error {
    if (backend == nil) {
        return fmt.Errorf("Database has not been opened.");
    }

    if (!util.PathExists(targetDir)) {
        err := util.MkDir(targetDir);
        if (err != nil) {
            return err;
        }
    }

    if (!util.IsEmptyDir(targetDir)) {
        return fmt.Errorf("Dump target dir '%s' is not empty.", targetDir);
    }

    return backend.DumpCourse(course, targetDir);
}

// Search the courses root directory and load all the associated courses and assignments.
// Return all the loaded course ids.
func LoadCourses() ([]string, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    return LoadCoursesFromDir(config.COURSES_ROOT.Get());
}

func MustLoadCourses() []string {
    courseIDs, err := LoadCourses();
    if (err != nil) {
        log.Fatal().Err(err).Str("path", config.COURSES_ROOT.Get()).Msg("Failed to load courses.");
    }

    return courseIDs;
}

func LoadCoursesFromDir(baseDir string) ([]string, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    log.Debug().Str("dir", baseDir).Msg("Searching for courses.");

    configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err);
    }

    log.Info().Int("count", len(configPaths)).Msg(fmt.Sprintf("Found %d course config(s).", len(configPaths)));

    courseIDs := make([]string, 0, len(configPaths));
    for _, configPath := range configPaths {
        courseID, err := LoadCourse(configPath);
        if (err != nil) {
            return nil, fmt.Errorf("Could not load course '%s': '%w'.", configPath, err);
        }

        courseIDs = append(courseIDs, courseID);
    }

    return courseIDs, nil;
}
