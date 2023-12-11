package db

import (
    "errors"
    "fmt"
    "path/filepath"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

func ClearCourse(course *model.Course) error {
    if (backend == nil) {
        return fmt.Errorf("Database has not been opened.");
    }

    return backend.ClearCourse(course);
}

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

// Load a course into the database from an existing path.
// This is meant for existing courses, for new courses use AddCourse().
func loadCourse(path string) (*model.Course, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    return backend.LoadCourse(path);
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

// Search the courses root directory and add all the associated courses and assignments.
// Return all the loaded course ids.
func AddCourses() ([]string, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    return AddCoursesFromDir(config.COURSES_ROOT.Get());
}

func MustAddCourses() []string {
    courseIDs, err := AddCourses();
    if (err != nil) {
        log.Fatal().Err(err).Str("path", config.COURSES_ROOT.Get()).Msg("Failed to load courses.");
    }

    return courseIDs;
}

func AddCoursesFromDir(baseDir string) ([]string, error) {
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
        course, err := AddCourse(configPath);
        if (err != nil) {
            return nil, fmt.Errorf("Could not load course '%s': '%w'.", configPath, err);
        }

        courseIDs = append(courseIDs, course.GetID());
    }

    return courseIDs, nil;
}

// Add a course to the db from the course's source.
func AddCourse(path string) (*model.Course, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    partialCourse, err := model.ReadCourseConfig(path);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to load course config '%s': '%w'.", path, err);
    }

    // If the source is empty, set it to this directory where it is being added from.
    if (partialCourse.Source.IsEmpty()) {
        partialCourse.Source = common.FileSpec(util.ShouldAbs(filepath.Dir(path)));
    }

    course, err := UpdateCourseFromSource(partialCourse);
    if (err != nil) {
        return nil, err;
    }

    if (course == nil) {
        return nil, fmt.Errorf("Course has no source.");
    }

    return course, nil;
}

func MustAddCourse(path string) *model.Course {
    course, err := AddCourse(path);
    if (err != nil) {
        log.Fatal().Err(err).Str("path", path).Msg("Failed to add course.");
    }

    return course;
}

// Get a fresh copy of the course from the source and load it into the DB
// (thereby updating the course).
// After the update, build new images.
func UpdateCourseFromSource(course *model.Course) (*model.Course, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    source := course.GetSource();

    if (source.IsEmpty() || source.IsNil()) {
        return nil, nil;
    }

    baseDir := course.GetBaseSourceDir();

    if (util.PathExists(baseDir)) {
        err := util.RemoveDirent(baseDir);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to remove existing course base source output '%s': '%w'.", baseDir, err);
        }
    }

    err := util.MkDir(baseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to make course base source dir '%s': '%w'.", baseDir, err);
    }

    err = source.CopyTarget(common.ShouldGetCWD(), baseDir, true);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to copy course source ('%s') into course base source dir ('%s'): '%w'.", source, baseDir, err);
    }

    configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err);
    }

    if (len(configPaths) == 0) {
        return nil, fmt.Errorf("Did not find any course configs in course source ('%s'), should be exactly one.", source);
    }

    if (len(configPaths) > 1) {
        return nil, fmt.Errorf("Found too many course configs (%d) in course source ('%s'), should be exactly one.", len(configPaths), source);
    }

    configPath := util.ShouldAbs(configPaths[0]);

    newCourse, err := loadCourse(configPath);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to load updated course: '%w'.", err);
    }

    // Ensure that the source is passed along.
    // This can happen when a course is loaded from a directory (without a source).
    if (newCourse.Source.IsEmpty()) {
        newCourse.Source = source;

        err = SaveCourse(newCourse);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to save new course: '%w'.", err);
        }
    }

    _, errs := newCourse.BuildAssignmentImages(false, false, docker.NewBuildOptions());
    if (len(errs) != 0) {
        err = nil;
        for _, newErr := range errs {
            err = errors.Join(err, newErr);
        }

        return nil, fmt.Errorf("Failed to build assignment images: '%w'.", err);
    }

    return newCourse, nil;
}
