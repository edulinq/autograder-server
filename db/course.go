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

    course, err := backend.LoadCourse(path);

    if (err == nil) {
        log.Info().Str("path", path).Str("id", course.GetID()).
                Int("num-assignments", len(course.Assignments)).Msg("Loaded course.");
    }

    return course, err;
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

    return AddCoursesFromDir(config.GetCourseImportDir(), nil);
}

func MustAddCourses() []string {
    courseIDs, err := AddCourses();
    if (err != nil) {
        log.Fatal().Err(err).Str("path", config.GetCourseImportDir()).Msg("Failed to load courses.");
    }

    return courseIDs;
}

func AddCoursesFromDir(baseDir string, source *common.FileSpec) ([]string, error) {
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
        course, err := AddCourse(configPath, source);
        if (err != nil) {
            return nil, fmt.Errorf("Could not load course '%s': '%w'.", configPath, err);
        }

        courseIDs = append(courseIDs, course.GetID());
    }

    return courseIDs, nil;
}

// Add a course to the db from a path.
func AddCourse(path string, source *common.FileSpec) (*model.Course, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    course, err := loadCourse(path);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to load course config '%s': '%w'.", path, err);
    }

    update := true;
    saveCourse := false;

    // Use the override source if not nil.
    if (source != nil) {
        course.Source = source;
        saveCourse = true;
    }

    // If the course's source is empty, set it to this directory where it is being added from.
    if ((course.Source == nil) || course.Source.IsEmpty()) {
        course.Source = common.GetPathFileSpec(util.ShouldAbs(filepath.Dir(path)));
        err = course.Source.Validate();
        if (err != nil) {
            return nil, fmt.Errorf("Failed to create source FileSpec: '%w'.", err);
        }

        saveCourse = true;
        update = false;
    }

    if (saveCourse) {
        err = SaveCourse(course);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to save course: '%w'.", err);
        }
    }

    if (!update) {
        return course, nil;
    }

    // Try to update the course from source.

    newCourse, _, err := UpdateCourseFromSource(course);
    if (err != nil) {
        return nil, err;
    }

    return newCourse, nil;
}

func MustAddCourse(path string) *model.Course {
    course, err := AddCourse(path, nil);
    if (err != nil) {
        log.Fatal().Err(err).Str("path", path).Msg("Failed to add course.");
    }

    return course;
}

// Get a fresh copy of the course from the source and load it into the DB
// (thereby updating the course).
// The new course (or old course if no update happens) will be returned.
// The boolean return indicates if an update attempt was made.
// Callers to this should consider if tasks should be stopped before,
// and if tasks should be started and images rebuilt after.
func UpdateCourseFromSource(course *model.Course) (*model.Course, bool, error) {
    if (backend == nil) {
        return nil, false, fmt.Errorf("Database has not been opened.");
    }

    source := course.GetSource();

    if ((source == nil) || source.IsEmpty() || source.IsNil()) {
        return course, false, nil;
    }

    baseDir := course.GetBaseSourceDir();

    if (util.PathExists(baseDir)) {
        err := util.RemoveDirent(baseDir);
        if (err != nil) {
            return nil, false, fmt.Errorf("Failed to remove existing course base source output '%s': '%w'.", baseDir, err);
        }
    }

    err := util.MkDir(baseDir);
    if (err != nil) {
        return nil, false, fmt.Errorf("Failed to make course base source dir '%s': '%w'.", baseDir, err);
    }

    err = source.CopyTarget(common.ShouldGetCWD(), baseDir, false);
    if (err != nil) {
        return nil, false, fmt.Errorf("Failed to copy course source ('%s') into course base source dir ('%s'): '%w'.", source, baseDir, err);
    }

    configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir);
    if (err != nil) {
        return nil, false, fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err);
    }

    if (len(configPaths) == 0) {
        return nil, false, fmt.Errorf("Did not find any course configs in course source ('%s'), should be exactly one.", source);
    }

    if (len(configPaths) > 1) {
        return nil, false, fmt.Errorf("Found too many course configs (%d) in course source ('%s'), should be exactly one.", len(configPaths), source);
    }

    configPath := util.ShouldAbs(configPaths[0]);

    newCourse, err := loadCourse(configPath);
    if (err != nil) {
        return nil, false, fmt.Errorf("Failed to load updated course: '%w'.", err);
    }

    // Ensure that the source is passed along.
    // This can happen when a course is loaded from a directory (without a source).
    if ((newCourse.Source == nil) || newCourse.Source.IsEmpty()) {
        newCourse.Source = source;

        err = SaveCourse(newCourse);
        if (err != nil) {
            return nil, false, fmt.Errorf("Failed to save new course: '%w'.", err);
        }
    }

    return newCourse, true, nil;
}
